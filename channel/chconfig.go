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

type AddrConf struct {
	Ip   string
	Port int
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
		ReadTimeout:  15,
		ReadBufSize:  10 * 1024,
		WriteTimeout: 15,
		WriteBufSize: 10 * 1024,
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
