/*
 * channel 相关配置
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

const (
	READ_TIMEOUT       = 20
	READ_BUFSIZE       = 128 * 1024
	WRITE_TIMEOUT      = 15
	WRITE_BUFSIZE      = 128 * 1024
	CLOSE_REV_FAILTIME = 3
)

// IChannelConf channel配置接口
type IChannelConf interface {
	// GetReadTimeout 获取读超时时间，单位为s
	GetReadTimeout() time.Duration

	// GetReadBufSize 获取读buffer大小，单位是字节(byte)
	GetReadBufSize() int

	// GetWriteTimeout 获取写超时时间，单位为s
	GetWriteTimeout() time.Duration

	// GetWriteBufSize 获取写buffer大小，单位是字节(byte)
	GetWriteBufSize() int

	// GetCloseRevFailTime 多少次读取失败后，关闭channel
	GetCloseRevFailTime() int

	// GetExtConfs 扩展配置
	GetExtConfs() map[string]interface{}

	// GetNetwork 获取通道协议类型
	// @see Network
	GetNetwork() Network

	// CopyChConf 复制配置
	CopyChConf(srcChConf IChannelConf)
}

// ChannelConf channel配置接口
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

	// 扩展配置
	ExtConfs map[string]interface{}
}

// NewChannelConf 创建配置
// readTimeout 读超时时间间隔，单位s
// readBufSize 读buffer大小，单位是字节(byte)
// writeTimeout 写超时时间间隔，单位s
// writeBufSize 写buffer大小，单位是字节(byte)
// network 使用的协议
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

// NewDefChannelConf 创建默认channelconf
// network 使用的协议
func NewDefChannelConf(network Network) *ChannelConf {
	return NewChannelConf(READ_TIMEOUT, READ_BUFSIZE, WRITE_TIMEOUT, WRITE_BUFSIZE, network)
}

// GetReadTimeout 获取读超时时间，单位为s
func (chConf *ChannelConf) GetReadTimeout() time.Duration {
	timeout := chConf.ReadTimeout
	if timeout <= 0 {
		timeout = READ_TIMEOUT
	}
	return timeout
}

// GetReadBufSize 获取读buffer大小，单位是字节(byte)
func (chConf *ChannelConf) GetReadBufSize() int {
	ret := chConf.ReadBufSize
	if ret <= 0 {
		ret = READ_BUFSIZE
	}
	return ret
}

// GetWriteTimeout 获取写超时时间，单位为s
func (chConf *ChannelConf) GetWriteTimeout() time.Duration {
	timeout := chConf.WriteTimeout
	if timeout <= 0 {
		timeout = WRITE_TIMEOUT
	}
	return timeout
}

// GetWriteBufSize 获取写buffer大小，单位是字节(byte)
func (chConf *ChannelConf) GetWriteBufSize() int {
	ret := chConf.WriteBufSize
	if ret <= 0 {
		ret = WRITE_BUFSIZE
	}
	return ret
}

// CloseRevFailTime 最大接收多少次失败后关闭
func (chConf *ChannelConf) GetCloseRevFailTime() int {
	ret := chConf.CloseRevFailTime
	if ret <= 0 {
		ret = CLOSE_REV_FAILTIME
	}
	return ret
}

// GetNetwork 获取通道协议类型
func (chConf *ChannelConf) GetNetwork() Network {
	return chConf.Network
}

// CopyChConf 复制配置
func (chConf *ChannelConf) CopyChConf(srcChConf IChannelConf) {
	chConf.ReadBufSize = srcChConf.GetReadBufSize()
	chConf.WriteTimeout = srcChConf.GetWriteTimeout()
	chConf.WriteBufSize = srcChConf.GetWriteBufSize()
	chConf.ReadTimeout = srcChConf.GetReadTimeout()
	chConf.CloseRevFailTime = srcChConf.GetCloseRevFailTime()
}

// GetExtConfs 扩展配置
func (chConf *ChannelConf) GetExtConfs() map[string]interface{} {
	return chConf.ExtConfs
}

// IAddrConf
type IAddrConf interface {
	// GetIp 获取ip或者url
	GetIp() string

	// GetPort 获取端口
	GetPort() int

	// GetAddrStr 获取完整地址
	GetAddrStr() string
}

// AddrConf
type AddrConf struct {
	Ip   string
	Port int
}

// NewAddrConf 地址配置
func NewAddrConf(ip string, port int) *AddrConf {
	b := &AddrConf{Ip: ip, Port: port}
	return b
}

// GetIp 获取ip
func (addr *AddrConf) GetIp() string {
	return addr.Ip
}

func (addr *AddrConf) GetPort() int {
	return addr.Port
}

// GetAddrStr 获取地址字符串
func (addr *AddrConf) GetAddrStr() string {
	if addr.Port <= 0 {
		return addr.Ip
	}
	return addr.Ip + fmt.Sprintf(":%v", addr.Port)
}

// ReadPoolConf 读资源池配置
type ReadPoolConf struct {
	MaxReadPoolSize  int
	MaxReadQueueSize int
}

// 每个队列最大读缓冲数
const MAX_READ_QUEUE_SIZE = 100

// 对应每个cpu的读线程池最大数
const MAX_READ_POOL_EVERY_CPU = 100

// NewReadPoolConf 初始化资源池
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

// NewDefReadPoolConf 创建默认的读线程池配置
func NewDefReadPoolConf() *ReadPoolConf {
	return &ReadPoolConf{
		MaxReadQueueSize: MAX_READ_QUEUE_SIZE,
		MaxReadPoolSize:  runtime.NumCPU() * MAX_READ_POOL_EVERY_CPU,
	}
}

var Global_Conf DefaultConf

// LoadDefaultConf 加载默认配置
func LoadDefaultConf() {
	Global_Conf = DefaultConf{
		ChannelConf:  channelConf,
		ReadPoolConf: readPoolConf}
	logx.Info("init global channelConf:", channelConf)
	logx.Info("init global readPoolConf:", channelConf)
}
