/*
 * Author:slive
 * DATE:2020/7/24
 */
package config

import (
	"fmt"
	"net/url"
	"os"
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
}

type LogConf struct {
	// LogFile 日志文件
	LogFile string

	// LogDir 日志路径
	LogDir string

	Level int
}

type AddrConf struct {
	Ip   string
	Port int
}

func (addr *AddrConf) GetAddrStr() string {
	return addr.Ip + fmt.Sprintf(":%v", addr.Port)
}

type ServerConf struct {
	AddrConf
	ChannelConf
	MaxAcceptSize int
}

type ClientConf struct {
	AddrConf
	ChannelConf
}

type KcpConf struct {
	// TODO kcp相关的配置
}

type KcpClientConf struct {
	ClientConf
	KcpConf
}

type KcpServerConf struct {
	ServerConf
	KcpConf
}

type UdpServerConf struct {
	ServerConf
}

type UdpClientConf struct {
	ClientConf
}

type HttpxServerConf struct {
	ServerConf
}

type HttpxClientConf struct {
	ClientConf
}

type WsClientConf struct {
	ClientConf
	Scheme      string
	SubProtocol []string
	Path        string
	Params      map[string]interface{}
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.Scheme, Host: wsClientConf.GetAddrStr(), Path: wsClientConf.Path}
	return u.String()
}

type ReadPoolConf struct {
	MaxReadPoolSize  int
	MaxReadQueueSize int
}

type GlobalConf struct {
	ChannelConf   *ChannelConf
	LogConf       *LogConf
	ReadPoolConf  *ReadPoolConf
	TcpServerConf *ServerConf
}

var (
	channelConf = &ChannelConf{
		ReadTimeout:  15,
		ReadBufSize:  10 * 1024,
		WriteTimeout: 15,
		WriteBufSize: 10 * 1024,
	}

	logConf = &LogConf{
		LogFile: "log-gsfly.log",
		LogDir:  getPwd() + "/log",
	}

	readPoolConf = &ReadPoolConf{
		MaxReadQueueSize: 100,
		MaxReadPoolSize:  runtime.NumCPU() * 10,
	}

	tcpServerConf = &ServerConf{
		AddrConf: AddrConf{
			Ip:   "127.0.0.1",
			Port: 9981,},
		MaxAcceptSize: 50000,
	}
)

var Global_Conf GlobalConf

func init() {
	Global_Conf = GlobalConf{
		ChannelConf:   channelConf,
		LogConf:       logConf,
		ReadPoolConf:  readPoolConf,
		TcpServerConf: tcpServerConf}
}

func getPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return pwd
}

func LoadConConf(path string) {
	// TODO
}
