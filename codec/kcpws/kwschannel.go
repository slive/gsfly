/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcpws

import (
	logx "github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
	gchannel "gsfly/channel"
	kcpx "gsfly/codec/kcp"
	"gsfly/codec/kcpws/frame"
	"gsfly/config"
)

type KwsChannel struct {
	kcpx.KcpChannel
	conn           *kcp.UDPSession
	handleKwsFrame HandleKwsFrame
}

func NewKwsChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf) *KwsChannel {
	ch := &KwsChannel{conn: kcpconn}
	ch.KcpChannel = *kcpx.NewKcpChannel(kcpconn, conf)
	return ch
}

func StartKwsChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf, handleKwsFrame HandleKwsFrame) *KwsChannel {
	ch := NewKwsChannel(kcpconn, conf)
	ch.handleKwsFrame = handleKwsFrame
	ch.SetHandleMsgFunc(handlerMessage)
	go gchannel.StartReadLoop(ch)
	return ch
}

func (b *KwsChannel) GetConn() *kcp.UDPSession {
	return b.conn
}

func (b *KwsChannel) Read() (packet gchannel.Packet, err error) {
	return kcpx.ReadKcp(b)
}

func (b *KwsChannel) Write(datapack gchannel.Packet) error {
	return kcpx.WriteKcp(b, datapack)
}

type KwsPacket struct {
	kcpx.KcpPacket
}

func (b *KwsChannel) NewPacket() gchannel.Packet {
	k := &KwsPacket{}
	k.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_KWS)
	return k
}

func handlerMessage(datapack gchannel.Packet) error {
	bf := frame.NewInputFrame(datapack.GetData())
	logx.Println("baseFrame:", bf.ToJsonString())
	conn := datapack.GetChannel()
	kwsConn := conn.(*KwsChannel)
	kwsConn.handleKwsFrame(conn, bf)
	return nil
}

type HandleKwsFrame func(conn gchannel.Channel, frame frame.Frame) error
