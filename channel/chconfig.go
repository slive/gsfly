/*
 * Author:slive
 * DATE:2020/7/24
 */
package channel

import (
	"fmt"
	"runtime"
	"time"
)

type ChannelConf struct {
	// ReadTimeout 读超时时间间隔，单位ms，默认15s
	ReadTimeout time.Duration

	// WriteTimeout 写超时时间间隔，单位ms，默认15s
	WriteTimeout time.Duration

	// ReadBufSize 读缓冲
	ReadBufSize int

	// ReadBufSize 读缓冲
	WriteBufSize int

	// 使用的协议
	Protocol Protocol
}

type IChannelConf interface {
	GetReadTimeout() time.Duration
	GetReadBufSize() int
	GetWriteTimeout() time.Duration
	GetWriteBufSize() int

	// GetProtocol 获取通道协议类型
	// @see Protocol
	GetProtocol() Protocol
}

const (
	READ_TIMEOUT  = 15
	READ_BUFSIZE  = 8 * 1024
	WRITE_TIMEOUT = 15
	WRITE_BUFSIZE = 8 * 1024
)

func NewChannelConf(readTimeout time.Duration, readBufSize int, writeTimeout time.Duration, writeBufSize int, protocol Protocol) IChannelConf {
	b := &ChannelConf{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ReadBufSize:  readBufSize,
		WriteBufSize: writeBufSize,
		Protocol:     protocol,
	}
	return b
}

func NewDefChannelConf(protocol Protocol) IChannelConf {
	return NewChannelConf(READ_TIMEOUT, READ_BUFSIZE, WRITE_TIMEOUT, WRITE_BUFSIZE, protocol)
}

func (bc *ChannelConf) GetReadTimeout() time.Duration {
	timeout := bc.ReadTimeout
	if timeout <= 0 {
		timeout = READ_TIMEOUT
	}
	return timeout
}

func (bc *ChannelConf) GetReadBufSize() int {
	ret := bc.ReadBufSize
	if ret <= 0 {
		ret = READ_BUFSIZE
	}
	return ret
}

func (bc *ChannelConf) GetWriteTimeout() time.Duration {
	timeout := bc.WriteTimeout
	if timeout <= 0 {
		timeout = WRITE_TIMEOUT
	}
	return timeout
}

func (bc *ChannelConf) GetWriteBufSize() int {
	ret := bc.WriteBufSize
	if ret <= 0 {
		ret = WRITE_BUFSIZE
	}
	return ret
}

// GetProtocol 获取通道协议类型
func (bc *ChannelConf) GetProtocol() Protocol {
	return bc.Protocol
}

type AddrConf struct {
	Ip   string
	Port int
}

type IAddrConf interface {
	// GetIp 获取ip或者url
	GetIp() string

	// GetPort 获取端口
	GetPort() int

	// GetAddrStr 获取完整地址
	GetAddrStr() string
}

func NewAddrConf(ip string, port int) IAddrConf {
	b := &AddrConf{Ip: ip, Port: port}
	return b
}

func (addr *AddrConf) GetIp() string {
	return addr.Ip
}

func (addr *AddrConf) GetPort() int {
	return addr.Port
}

func (addr *AddrConf) GetAddrStr() string {
	if addr.Port <= 0 {
		return addr.Ip
	}
	return addr.Ip + fmt.Sprintf(":%v", addr.Port)
}

type ReadPoolConf struct {
	MaxReadPoolSize  int
	MaxReadQueueSize int
}

type DefaultConf struct {
	ChannelConf  *ChannelConf
	ReadPoolConf *ReadPoolConf
}

var (
	channelConf = &ChannelConf{
		ReadTimeout:  READ_TIMEOUT,
		ReadBufSize:  READ_BUFSIZE,
		WriteTimeout: WRITE_TIMEOUT,
		WriteBufSize: WRITE_BUFSIZE,
	}

	readPoolConf = &ReadPoolConf{
		MaxReadQueueSize: 100,
		MaxReadPoolSize:  runtime.NumCPU() * 10,
	}
)

var Global_Conf DefaultConf

func LoadDefaultConf() {
	Global_Conf = DefaultConf{
		ChannelConf:  channelConf,
		ReadPoolConf: readPoolConf}
}

func LoadConConf(path string) {
	// TODO
}
