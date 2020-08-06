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

type BaseChannelConf struct {
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

type ChannelConf interface {
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
	READ_BUFSIZE  = 10 * 1024
	WRITE_TIMEOUT = 15
	WRITE_BUFSIZE = 10 * 1024
)

func NewBaseChannelConf(readTimeout time.Duration, readBufSize int, writeTimeout time.Duration, writeBufSize int, protocol Protocol) ChannelConf {
	b := &BaseChannelConf{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ReadBufSize:  readBufSize,
		WriteBufSize: writeBufSize,
		Protocol:     protocol,
	}
	return b
}

func NewDefaultChannelConf() ChannelConf {
	return NewBaseChannelConf(READ_TIMEOUT, READ_BUFSIZE, WRITE_TIMEOUT, WRITE_BUFSIZE, PROTOCOL_TCP)
}

func (bc *BaseChannelConf) GetReadTimeout() time.Duration {
	timeout := bc.ReadTimeout
	if timeout <= 0 {
		timeout = READ_TIMEOUT
	}
	return timeout
}

func (bc *BaseChannelConf) GetReadBufSize() int {
	ret := bc.ReadBufSize
	if ret <= 0 {
		ret = READ_BUFSIZE
	}
	return ret
}

func (bc *BaseChannelConf) GetWriteTimeout() time.Duration {
	timeout := bc.WriteTimeout
	if timeout <= 0 {
		timeout = WRITE_TIMEOUT
	}
	return timeout
}

func (bc *BaseChannelConf) GetWriteBufSize() int {
	ret := bc.WriteBufSize
	if ret <= 0 {
		ret = WRITE_BUFSIZE
	}
	return ret
}

// GetProtocol 获取通道协议类型
func (bc *BaseChannelConf) GetProtocol() Protocol {
	return bc.Protocol
}

type BaseAddrConf struct {
	Ip   string
	Port int
	Protocol Protocol
}

type AddrConf interface {
	// GetIp 获取ip或者url
	GetIp() string

	// GetPort 获取端口
	GetPort() int

	// GetAddrStr 获取完整地址
	GetAddrStr() string
}

func NewBaseAddrConf(ip string, port int) AddrConf {
	b := &BaseAddrConf{Ip: ip, Port: port}
	return b
}

func (addr *BaseAddrConf) GetIp() string {
	return addr.Ip
}

func (addr *BaseAddrConf) GetPort() int {
	return addr.Port
}

func (addr *BaseAddrConf) GetAddrStr() string {
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
	ChannelConf  *BaseChannelConf
	ReadPoolConf *ReadPoolConf
}

var (
	channelConf = &BaseChannelConf{
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
