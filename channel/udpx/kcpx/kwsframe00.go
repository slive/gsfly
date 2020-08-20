/*
 * kcp-ws转化桢
 * Author:slive
 * DATE:2020/7/7
 */
package kcpx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	logx "github.com/Slive/gsfly/logger"
	"log"
)

type Frame interface {
	Unpack()

	Pack()

	GetKcpData() (kcpData []byte)

	GetPayload() (payload []byte)

	GetRemoteIp() (remoteIp string)

	GetPort() (port uint16)

	GetOpCode() (opCode uint16)

	ToJsonString() string
}

const (
	VERSION_00             = uint16(0x00)
	OPCODE_TEXT_SESSION    = uint16(0x01)
	OPCODE_TEXT_SIGNALLING = uint16(0X02)
	OPCODE_CLOSE           = uint16(0X08)
	OPCODE_PING            = uint16(0X09)
	OPCODE_PONG            = uint16(0X0A)
)

// BaseFrame 基于kcp数据之上定义的数据帧
//
// 包括格式如：
//
type BaseFrame struct {
	// KcpData 用于kcp底层通信发生的数据
	KcpData      []byte
	Version      uint16
	OpCode       uint16
	PayloadLen   uint32
	Payload      []byte
	IpFlag       int8
	RemoteIpData []byte
	RemoteIp     string
	Port         uint16
	ReadOnly     bool
}

func (baseFrame *BaseFrame) Pack() {
	defer func() {
		rc := recover()
		if rc != nil {
			err := errors.New("decode error")
			log.Println(err)
			return
		}
	}()
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, baseFrame.Version)
	binary.Write(buf, binary.BigEndian, baseFrame.OpCode)
	binary.Write(buf, binary.BigEndian, baseFrame.PayloadLen)
	binary.Write(buf, binary.BigEndian, baseFrame.Payload)
	ipData := baseFrame.RemoteIpData
	if len(ipData) > 0 {
		binary.Write(buf, binary.BigEndian, baseFrame.IpFlag)
		binary.Write(buf, binary.BigEndian, ipData)
		binary.Write(buf, binary.BigEndian, baseFrame.Port)
	}

	// kcp通信层将kcpdata数据发送
	baseFrame.KcpData = buf.Bytes()
}

func (baseFrame *BaseFrame) Unpack() {
	defer func() {
		err := recover()
		if err != nil {
			logx.Error("encode error:", err)
		}
	}()

	kcpdata := baseFrame.KcpData
	kcpLen := len(kcpdata)
	if kcpdata == nil || kcpLen == 0 {
		logx.Info("kcpdata is empty.")
		return
	}
	baseFrame.ReadOnly = true

	// 记录数据解析的索引
	var index uint32 = 0

	// 解析version
	baseFrame.Version = binary.BigEndian.Uint16(kcpdata[index : index+2])
	index += 2

	// 解析opcode
	baseFrame.OpCode = binary.BigEndian.Uint16(kcpdata[index : index+2])
	index += 2

	// 解析playloadlen
	var palyloadLen uint32 = binary.BigEndian.Uint32(kcpdata[index : index+4])
	baseFrame.PayloadLen = palyloadLen
	index += 4

	// 解析playload，数据长度由palyloadlen确定
	palyloadEnd := index + palyloadLen
	baseFrame.Payload = kcpdata[index:palyloadEnd]
	index = palyloadEnd

	// 可选操作
	ipMaxIndex := index + 1 + 4 + 2
	if kcpLen > int(ipMaxIndex) {
		// 解析远端ip，记录远端ip的目的是为了解决udp协议的nat问题
		var remoteIp string
		var ipFlag int8 = int8(kcpdata[index])
		index += 1

		remoteIpData := kcpdata[index : index+4]
		if ipFlag == 0 {
			// ipv4
			remoteIp = fmt.Sprintf("%d.%d.%d.%d", remoteIpData[0], remoteIpData[1], remoteIpData[2], remoteIpData[3])
		} else {
			// ipv6
			remoteIp = fmt.Sprintf("%A:%A:%A:%A", remoteIpData[0], remoteIpData[1], remoteIpData[2], remoteIpData[3])
		}
		index += 4

		// 解析端口
		var port uint16 = binary.BigEndian.Uint16(kcpdata[index : index+2])
		index += 2

		baseFrame.IpFlag = ipFlag
		baseFrame.RemoteIpData = remoteIpData
		baseFrame.RemoteIp = remoteIp
		baseFrame.Port = port
	}
}

func FetchOpcode(kcpData []byte) uint16 {
	return binary.BigEndian.Uint16(kcpData[2:4])
}

func (baseFrame *BaseFrame) GetKcpData() (kcpData []byte) {
	return baseFrame.KcpData
}

func (baseFrame *BaseFrame) GetPayload() (payload []byte) {
	return baseFrame.Payload
}

func (baseFrame *BaseFrame) GetRemoteIp() (remoteIp string) {
	return baseFrame.RemoteIp
}

func (baseFrame *BaseFrame) GetPort() (port uint16) {
	return baseFrame.Port
}

func (baseFrame *BaseFrame) ToJsonString() string {
	return string(baseFrame.Payload)
}

func (baseFrame *BaseFrame) GetOpCode() (opCode uint16) {
	return baseFrame.OpCode
}

type TextFrame struct {
	BaseFrame
	PayloadStr string `json:"payloadStr"`
}

func (frame *TextFrame) Unpack() {
	frame.BaseFrame.Unpack()
	frame.PayloadStr = string(frame.Payload)
}

func (frame *TextFrame) Pack() {
	str := frame.PayloadStr
	if len(str) > 0 {
		frame.Payload = []byte(str)
	}
	frame.BaseFrame.Pack()
}

type TextSessionFrame struct {
	TextFrame
}

func (frame *TextSessionFrame) Unpack() {
	frame.TextFrame.Unpack()
}

func (frame *TextSessionFrame) Pack() {
	frame.OpCode = OPCODE_TEXT_SESSION
	frame.TextFrame.Pack()
}

type TextSignallingFrame struct {
	TextFrame
}

func (frame *TextSignallingFrame) Unpack() {
	frame.TextFrame.Unpack()
}

func (frame *TextSignallingFrame) Pack() {
	frame.OpCode = OPCODE_TEXT_SIGNALLING
	frame.TextFrame.Pack()
}

func NewInputFrame(kcpData []byte) Frame {
	opCode := binary.BigEndian.Uint16(kcpData[2:4])
	logx.Debug("opCode:", opCode)
	if OPCODE_TEXT_SESSION == opCode {
		t := &TextSessionFrame{TextFrame{
			BaseFrame: BaseFrame{KcpData: kcpData},
		}}
		t.Unpack()
		return t
	} else if OPCODE_TEXT_SIGNALLING == opCode {
		t := &TextSignallingFrame{TextFrame{
			BaseFrame: BaseFrame{KcpData: kcpData},
		}}
		t.Unpack()
		return t
	} else {
		t := &TextFrame{
			BaseFrame: BaseFrame{KcpData: kcpData},
		}
		t.Unpack()
		return t
	}
}

func NewOutputFrame(opCode uint16, payload []byte) Frame {
	if OPCODE_TEXT_SESSION == opCode {
		t := &TextSessionFrame{TextFrame{
			BaseFrame: BaseFrame{Payload: payload,
				PayloadLen: uint32(len(payload)),
				OpCode:     opCode},
		}}
		t.Pack()
		return t
	} else if OPCODE_TEXT_SIGNALLING == opCode {
		t := &TextSignallingFrame{TextFrame{
			BaseFrame: BaseFrame{Payload: payload,
				PayloadLen: uint32(len(payload)),
				OpCode:     opCode},
		}}
		t.Pack()
		return t
	} else {
		t := &TextFrame{
			BaseFrame: BaseFrame{Payload: payload,
				PayloadLen: uint32(len(payload)),
				OpCode:     opCode}}
		t.Pack()
		return t
	}
}
