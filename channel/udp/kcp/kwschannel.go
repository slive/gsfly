/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"github.com/xtaci/kcp-go"
	gchannel "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
)

type KwsChannel struct {
	KcpChannel
	conn           *kcp.UDPSession
	handleKwsFrame HandleKwsFrame
}

func NewKwsChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf) *KwsChannel {
	ch := &KwsChannel{conn: kcpconn}
	ch.KcpChannel = *NewKcpChannel(kcpconn, conf)
	return ch
}

func StartKwsChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf, handleKwsFrame HandleKwsFrame) (*KwsChannel, error) {
	return StartKwsChannelWithHandle(kcpconn, conf, handleKwsFrame, nil, nil)
}

func StartKwsChannelWithHandle(kcpconn *kcp.UDPSession, conf *config.ChannelConf, handleKwsFrame HandleKwsFrame, handleStartFunc gchannel.HandleStartFunc,
	handleCloseFunc gchannel.HandleCloseFunc) (*KwsChannel, error) {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start kwschannel error:", re)
		}
	}()
	var err error
	ch := NewKwsChannel(kcpconn, conf)
	ch.handleKwsFrame = handleKwsFrame
	ch.SetHandleMsgFunc(handlerMessage)
	go gchannel.StartReadLoop(ch)
	handle := ch.ChannelHandle
	startFunc := handle.HandleStartFunc
	if startFunc != nil {
		err = startFunc(ch)
	}
	handle.HandleStartFunc = handleStartFunc
	handle.HandleCloseFunc = handleCloseFunc
	return ch, err
}

func (b *KwsChannel) GetConn() *kcp.UDPSession {
	return b.conn
}

func (b *KwsChannel) Read() (packet gchannel.Packet, err error) {
	return ReadKcp(b)
}

func (b *KwsChannel) Write(datapack gchannel.Packet) error {
	return WriteKcp(b, datapack)
}

type KwsPacket struct {
	KcpPacket
}

func (b *KwsChannel) NewPacket() gchannel.Packet {
	k := &KwsPacket{}
	k.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_KWS)
	return k
}

func handlerMessage(datapack gchannel.Packet) error {
	bf := NewInputFrame(datapack.GetData())
	logx.Info("baseFrame:", bf.ToJsonString())
	conn := datapack.GetChannel()
	kwsConn := conn.(*KwsChannel)
	kwsConn.handleKwsFrame(conn, bf)
	return nil
}

type HandleKwsFrame func(conn gchannel.Channel, frame Frame) error
