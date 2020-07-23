/*
 * Author:slive
 * DATE:2020/7/7
 */
package frame

import "encoding/binary"

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
		return &TextFrame{
			BaseFrame: BaseFrame{KcpData: kcpData},
		}
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
