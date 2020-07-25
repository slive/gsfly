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

	SetHandleMsgFunc(handleMsgFunc HandleMsgFunc)

	GetHandleMsgFunc() HandleMsgFunc

	GetConf() *config.ChannelConf

	Close()
}

type HandleMsgFunc func(packet Packet) error

type BaseChannel struct {
	handleMsgFunc HandleMsgFunc
	conf          *config.ChannelConf
	closeExit     chan bool
}

var readPoolConf = config.Global_Conf.ReadPoolConf

var readPool *ReadPool = NewReadPool(readPoolConf.MaxReadPoolSize, readPoolConf.MaxReadQueueSize)

func NewBaseChannel(conf *config.ChannelConf) *BaseChannel {
	conn := &BaseChannel{conf: conf, closeExit: make(chan bool, 1)}
	return conn
}

func (b *BaseChannel) GetConf() *config.ChannelConf {
	return b.conf
}

func (b *BaseChannel) SetHandleMsgFunc(handleMsgFunc HandleMsgFunc) {
	b.handleMsgFunc = handleMsgFunc
}

func (b *BaseChannel) GetHandleMsgFunc() HandleMsgFunc {
	return b.handleMsgFunc
}

func (b *BaseChannel) Close() {
	b.closeExit <- true
	close(b.closeExit)
}

func StartReadLoop(channel Channel) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("error:", i)
		}
	}()
	for {
		rev, err := channel.Read()
		if err != nil {
			return err
		}

		if rev.GetPrepare() {
			readPool.Cache(rev)
		}
	}
}
