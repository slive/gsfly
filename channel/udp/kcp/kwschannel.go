/*
 * kcp-ws通信通道
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

func newKwsChannel(kcpConn *kcp.UDPSession, conf *config.ChannelConf) *KwsChannel {
	ch := &KwsChannel{conn: kcpConn}
	ch.KcpChannel = *newKcpChannel(kcpConn, conf)
	return ch
}

func NewKwsChannel(kcpConn *kcp.UDPSession, conf *config.ChannelConf, handleKwsFrame HandleKwsFrame) *KwsChannel {
	return NewKwsChannelWithHandle(kcpConn, conf, handleKwsFrame, nil, nil)
}

func NewKwsChannelWithHandle(kcpConn *kcp.UDPSession, conf *config.ChannelConf, handleKwsFrame HandleKwsFrame, handleStartFunc gchannel.HandleStartFunc,
	handleCloseFunc gchannel.HandleCloseFunc) *KwsChannel {
	ch := newKwsChannel(kcpConn, conf)
	ch.handleKwsFrame = handleKwsFrame
	ch.SetHandleMsgFunc(handlerMessage)
	handle := ch.ChannelHandle
	handle.HandleStartFunc = handleStartFunc
	handle.HandleCloseFunc = handleCloseFunc
	return ch
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

// handlerMessage 实现消息处理，并调用HandleKwsFrame
func handlerMessage(datapack gchannel.Packet) error {
	bf := NewInputFrame(datapack.GetData())
	logx.Info("baseFrame:", bf.ToJsonString())
	conn := datapack.GetChannel()
	kwsConn := conn.(*KwsChannel)
	kwsConn.handleKwsFrame(conn, bf)
	return nil
}

// HandleKwsFrame 处理kwsframe方法
type HandleKwsFrame func(conn gchannel.Channel, frame Frame) error
