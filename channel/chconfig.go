/*
 * Author:slive
 * DATE:2020/7/24
 */
package channel

import (
	"fmt"
	logx "github.com/Slive/gsfly/logger"
	"runtime"
	"time"
)

type ChannelConf struct {
	// ReadTimeout 读超时时间间隔，单位s
	ReadTimeout time.Duration

	// WriteTimeout 写超时时间间隔，单位s
	WriteTimeout time.Duration

	// ReadBufSize 读缓冲
	ReadBufSize int

	// ReadBufSize 读缓冲
	WriteBufSize int

	// CloseRevFailTime 最大接收多少次失败后关闭
	CloseRevFailTime int

	// 使用的协议
	Network Network

	ExtConfs map[string]interface{}
}

type IChannelConf interface {
	GetReadTimeout() time.Duration
	GetReadBufSize() int
	GetWriteTimeout() time.Duration
	GetWriteBufSize() int
	GetCloseRevFailTime() int

	// GetExtConfs 扩展配置
	GetExtConfs() map[string]interface{}

	// GetNetwork 获取通道协议类型
	// @see Network
	GetNetwork() Network
}

const (
	READ_TIMEOUT       = 20
	READ_BUFSIZE       = 128 * 1024
	WRITE_TIMEOUT      = 15
	WRITE_BUFSIZE      = 128 * 1024
	CLOSE_REV_FAILTIME = 3
)

func NewChannelConf(readTimeout time.Duration, readBufSize int, writeTimeout time.Duration,
	writeBufSize int, network Network) *ChannelConf {
	b := &ChannelConf{
		ReadTimeout:      readTimeout,
		WriteTimeout:     writeTimeout,
		ReadBufSize:      readBufSize,
		WriteBufSize:     writeBufSize,
		Network:          network,
		CloseRevFailTime: CLOSE_REV_FAILTIME,
	}
	return b
}

func NewDefChannelConf(network Network) *ChannelConf {
	return NewChannelConf(READ_TIMEOUT, READ_BUFSIZE, WRITE_TIMEOUT, WRITE_BUFSIZE, network)
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

func (bc *ChannelConf) GetCloseRevFailTime() int {
	ret := bc.CloseRevFailTime
	if ret <= 0 {
		ret = CLOSE_REV_FAILTIME
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

// GetNetwork 获取通道协议类型
func (bc *ChannelConf) GetNetwork() Network {
	return bc.Network
}

// GetExtConfs 扩展配置
func (bc *ChannelConf) GetExtConfs() map[string]interface{} {
	return bc.ExtConfs
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

func NewAddrConf(ip string, port int) *AddrConf {
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

const MAX_READ_QUEUE_SIZE = 100
const MAX_READ_POOL_EVERY_CPU = 100

func NewReadPoolConf(maxReadPoolSize, maxReadQueueSize int) *ReadPoolConf {
	r := &ReadPoolConf{
		MaxReadQueueSize: maxReadQueueSize,
		MaxReadPoolSize:  maxReadPoolSize,
	}
	return r
}

type DefaultConf struct {
	ChannelConf  *ChannelConf
	ReadPoolConf *ReadPoolConf
}

func init() {
	LoadDefaultConf()
}

var (
	channelConf = &ChannelConf{
		ReadTimeout:      READ_TIMEOUT,
		ReadBufSize:      READ_BUFSIZE,
		WriteTimeout:     WRITE_TIMEOUT,
		WriteBufSize:     WRITE_BUFSIZE,
		CloseRevFailTime: CLOSE_REV_FAILTIME,
	}
	readPoolConf = NewDefReadPoolConf()
)

func NewDefReadPoolConf() *ReadPoolConf {
	return &ReadPoolConf{
		MaxReadQueueSize: MAX_READ_QUEUE_SIZE,
		MaxReadPoolSize:  runtime.NumCPU() * MAX_READ_POOL_EVERY_CPU,
	}
}

var Global_Conf DefaultConf

func LoadDefaultConf() {
	Global_Conf = DefaultConf{
		ChannelConf:  channelConf,
		ReadPoolConf: readPoolConf}
	logx.Info("init global channelConf:", channelConf)
	logx.Info("init global readPoolConf:", channelConf)
}
