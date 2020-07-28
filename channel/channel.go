/*
 * 通信连接
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
	// 创建收发包
	NewPacket() Packet

	// 通道Id
	GetChId() string

	// 读取方法
	Read() (packet Packet, err error)

	// 写入方法
	Write(packet Packet) error

	// 处理读取消息方法的方法
	GetHandleMsgFunc() HandleMsgFunc

	// 获取通道配置
	GetConf() *config.ChannelConf

	// 获取通道统计相关
	GetStatis() *Statis

	// 关闭
	Close()
}

// ChannelHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	HandleStartFunc HandleStartFunc
	HandleMsgFunc   HandleMsgFunc
	HandleCloseFunc HandleCloseFunc
}
type HandleMsgFunc func(packet Packet) error

type HandleCloseFunc func(channel Channel) error

type HandleStartFunc func(channel Channel) error

// Statis 统计相关，比如收发包数目，收发次数
type Statis struct {
	RevByteNum    int64
	RevPacketNum  int64
	RevTime       time.Time
	SendByteNum   int64
	SendPacketNum int64
	SendTime      time.Time
}

type BaseChannel struct {
	ChannelHandle
	Statis
	Conf      *config.ChannelConf
	closeExit chan bool
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

var readPoolConf = config.Global_Conf.ReadPoolConf

var readPool *ReadPool = NewReadPool(readPoolConf.MaxReadPoolSize, readPoolConf.MaxReadQueueSize)

func NewBaseChannel(conf *config.ChannelConf) *BaseChannel {
	conn := &BaseChannel{Conf: conf, closeExit: make(chan bool, 1), ChannelHandle: ChannelHandle{}, Statis: Statis{}}
	return conn
}

func (b *BaseChannel) GetConf() *config.ChannelConf {
	return b.Conf
}

func (b *BaseChannel) GetStatis() *Statis {
	return &b.Statis
}

func (b *BaseChannel) SetHandleMsgFunc(handleMsgFunc HandleMsgFunc) {
	b.ChannelHandle.HandleMsgFunc = handleMsgFunc
}

func (b *BaseChannel) GetHandleMsgFunc() HandleMsgFunc {
	return b.ChannelHandle.HandleMsgFunc
}

func (b *BaseChannel) Close() {
	b.closeExit <- true
	close(b.closeExit)
	// TODO 各自处理？
	closeFunc := b.HandleCloseFunc
	if closeFunc != nil {
		closeFunc(b)
	}
}

func StartReadLoop(channel Channel) error {
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

		if rev != nil && rev.GetPrepare() {
			readPool.Cache(rev)
		}
	}
}

func RevStatis(packet Packet) {
	statis := packet.GetChannel().GetStatis()
	statis.RevByteNum += int64(len(packet.GetData()))
	statis.RevPacketNum += 1
	statis.RevTime = time.Now()
}

func SendStatis(packet Packet) {
	statis := packet.GetChannel().GetStatis()
	statis.SendByteNum += int64(len(packet.GetData()))
	statis.SendPacketNum += 1
	statis.SendTime = time.Now()
}
