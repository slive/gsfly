/*
 * 通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package channel

import (
	"gsfly/config"
	logx "gsfly/logger"
	"time"
)

// Channel 通信通道接口
type Channel interface {

	// 启动通信通道
	StartChannel(channel Channel) error

	// NewPacket 创建收发包
	NewPacket() Packet

	// GetChId 通道Id
	GetChId() string

	// Read 读取方法
	Read() (packet Packet, err error)

	// Write 写入方法
	Write(packet Packet) error

	// GetHandleMsgFunc 处理读取消息方法的方法
	GetHandleMsgFunc() HandleMsgFunc

	// GetChConf 获取通道配置
	GetChConf() *config.ChannelConf

	// GetChStatis 获取通道统计相关
	GetChStatis() *ChannelStatis

	// 停止通道
	StopChannel()
}

// ChannelHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	HandleStartFunc HandleStartFunc
	HandleMsgFunc   HandleMsgFunc
	HandleCloseFunc HandleCloseFunc
}

// HandleMsgFunc 处理消息方法
type HandleMsgFunc func(packet Packet) error

// HandleCloseFunc 处理关闭时的方法
type HandleCloseFunc func(channel Channel) error

// HandleStartFunc 处理启动时的方法
type HandleStartFunc func(channel Channel) error

// NewChHandle 创建处理方法类
func NewChHandle(handleMsgFunc HandleMsgFunc, handleStartFunc HandleStartFunc, handleCloseFunc HandleCloseFunc) *ChannelHandle {
	return &ChannelHandle{
		HandleStartFunc: handleStartFunc,
		HandleMsgFunc:   handleMsgFunc,
		HandleCloseFunc: handleCloseFunc,
	}
}

// ChannelStatis 统计相关，比如收发包数目，收发次数
type ChannelStatis struct {
	RevByteNum    int64
	RevPacketNum  int64
	RevTime       time.Time
	SendByteNum   int64
	SendPacketNum int64
	SendTime      time.Time
}

func NewChStatis() *ChannelStatis {
	return &ChannelStatis{
		RevByteNum:    0,
		RevPacketNum:  0,
		RevTime:       time.Now(),
		SendByteNum:   0,
		SendPacketNum: 0,
		SendTime:      time.Now(),
	}
}

// BaseChannel channel基类
type BaseChannel struct {
	ChannelHandle
	ChannelStatis
	chConf    *config.ChannelConf
	readPool  *ReadPool
	closeExit chan bool
}

var Default_Read_Pool_Conf = config.Global_Conf.ReadPoolConf

// 初始化读协程池，全局配置
var Default_ReadPool *ReadPool = NewReadPool(Default_Read_Pool_Conf.MaxReadPoolSize, Default_Read_Pool_Conf.MaxReadQueueSize)

// NewDefaultBaseChannel 创建默认基础通信通道
func NewDefaultBaseChannel(conf *config.ChannelConf) *BaseChannel {
	return NewBaseChannel(conf, Default_ReadPool)
}

func NewBaseChannel(conf *config.ChannelConf, readPool *ReadPool) *BaseChannel {
	conn := &BaseChannel{chConf: conf, closeExit: make(chan bool, 1), ChannelHandle: ChannelHandle{}, ChannelStatis: *NewChStatis(), readPool: readPool}
	return conn
}

func (b *BaseChannel) StartChannel(channel Channel) error {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start updchannel error:", re)
		}
	}()
	go b.startReadLoop(channel)

	// 启动后处理方法
	startFunc := b.HandleStartFunc
	if startFunc != nil {
		err := startFunc(channel)
		if err != nil {
			b.StopChannel()
		}
		return err
	}
	logx.Info("finish to start channel, chId:", channel.GetChId())
	return nil
}

func (b *BaseChannel) NewPacket() Packet {
	panic("implement me")
}

func (b *BaseChannel) GetChId() string {
	panic("implement me")
}

func (b *BaseChannel) Read() (packet Packet, err error) {
	panic("implement me")
}

func (b *BaseChannel) Write(packet Packet) error {
	panic("implement me")
}

func (b *BaseChannel) GetChConf() *config.ChannelConf {
	return b.chConf
}

func (b *BaseChannel) GetChStatis() *ChannelStatis {
	return &b.ChannelStatis
}

func (b *BaseChannel) SetHandleMsgFunc(handleMsgFunc HandleMsgFunc) {
	b.ChannelHandle.HandleMsgFunc = handleMsgFunc
}

func (b *BaseChannel) GetHandleMsgFunc() HandleMsgFunc {
	return b.ChannelHandle.HandleMsgFunc
}

func (b *BaseChannel) StopChannel() {
	b.closeExit <- true
	close(b.closeExit)

	// TODO 各自处理？
	closeFunc := b.HandleCloseFunc
	if closeFunc != nil {
		closeFunc(b)
	}
}

// StartReadLoop 启动循环读取，读取到数据包后，放入#ReadQueue中，等待处理
func (b *BaseChannel) startReadLoop(channel Channel) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("read loop error:", i)
		}
	}()
	for {
		rev, err := channel.Read()
		if err != nil {
			return err
		}

		if rev != nil && rev.IsPrepare() {
			readPool := b.readPool
			if readPool != nil {
				// 放入读取协程池等待处理
				readPool.Cache(rev)
			} else {
				// 否则默认直接处理
				msgFunc := channel.GetHandleMsgFunc()
				if msgFunc != nil {
					msgFunc(rev)
				}
			}
		}
	}
}

// 读取统计
func RevStatis(packet Packet) {
	statis := packet.GetChannel().GetChStatis()
	statis.RevByteNum += int64(len(packet.GetData()))
	statis.RevPacketNum += 1
	statis.RevTime = time.Now()
}

// 写统计
func SendStatis(packet Packet) {
	statis := packet.GetChannel().GetChStatis()
	statis.SendByteNum += int64(len(packet.GetData()))
	statis.SendPacketNum += 1
	statis.SendTime = time.Now()
}
