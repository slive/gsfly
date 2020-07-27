/*
 * 通信连接
 * Author:slive
 * DATE:2020/7/17
 */
package channel

import (
	"gsfly/config"
	logx "gsfly/logger"
)

type Channel interface {
	NewPacket() Packet

	GetChId() string

	Read() (packet Packet, err error)

	Write(packet Packet) error

	GetHandleMsgFunc() HandleMsgFunc

	GetConf() *config.ChannelConf

	Close()
}

type ChannelHandle struct {
	HandleStartFunc HandleStartFunc
	HandleMsgFunc   HandleMsgFunc
	HandleCloseFunc HandleCloseFunc
}

type HandleMsgFunc func(packet Packet) error

type HandleCloseFunc func(channel Channel) error

type HandleStartFunc func(channel Channel) error

type BaseChannel struct {
	ChannelHandle
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
	conn := &BaseChannel{Conf: conf, closeExit: make(chan bool, 1), ChannelHandle: ChannelHandle{}}
	return conn
}

func (b *BaseChannel) GetConf() *config.ChannelConf {
	return b.Conf
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
	if closeFunc != nil{
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
