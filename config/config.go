/*
 * Author:slive
 * DATE:2020/7/24
 */
package config

import (
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
}

type LogConf struct {
	// LogFile 日志文件
	LogFile string

	// LogDir 日志路径
	LogDir string
}

type ReadPoolConf struct {
	MaxReadPoolSize  int
	MaxReadQueueSize int
}

type GlobalConf struct {
	ChannelConf  *ChannelConf
	LogConf      *LogConf
	ReadPoolConf *ReadPoolConf
}

var (
	channelConf = &ChannelConf{
		ReadTimeout:  15,
		ReadBufSize: 10*1024,
		WriteTimeout: 15,
	}

	logConf = &LogConf{
		LogFile: "log-gsfly.log",
		LogDir:  getPwd() + "/log",
	}

	readPoolConf = &ReadPoolConf{
		MaxReadQueueSize: 100,
		MaxReadPoolSize:  runtime.NumCPU() * 10,
	}
)

var Global_Conf GlobalConf

func init() {
	Global_Conf = GlobalConf{ChannelConf: channelConf,
		LogConf:      logConf,
		ReadPoolConf: readPoolConf}
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
