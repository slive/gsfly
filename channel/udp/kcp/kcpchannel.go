/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type KcpChannel struct {
	gch.BaseChannel
	conn     *kcp.UDPSession
	protocol gch.Protocol
}

func newKcpChannel(kcpConn *kcp.UDPSession, conf *gch.BaseChannelConf, protocol gch.Protocol) *KcpChannel {
	ch := &KcpChannel{conn: kcpConn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	ch.protocol = protocol
	readBufSize := conf.GetReadBufSize()
	kcpConn.SetReadBuffer(readBufSize)
	writeBufSize := conf.GetWriteBufSize()
	kcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewKcpChannel(kcpConn *kcp.UDPSession, chConf *gch.BaseChannelConf, msgFunc gch.HandleMsgFunc) *KcpChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewKcpChannelWithHandle(kcpConn, chConf, chHandle)
}

func NewKcpChannelWithHandle(kcpConn *kcp.UDPSession, chConf *gch.BaseChannelConf, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := newKcpChannel(kcpConn, chConf, chConf.Protocol)
	ch.ChannelHandle = *chHandle
	ch.SetChId("kcp-" + kcpConn.LocalAddr().String() + "-" + kcpConn.RemoteAddr().String() + "-" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

func (b *KcpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conn := b.conn
	conf := b.GetChConf()
	// conn.SetReadDeadline(time.Now().Add(conf.GetReadTimeout() * time.Second))
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
		logx.Info("read kcp err:", err)
		return nil, err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil, nil
	}

	datapack := b.NewPacket()
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	gch.RevStatis(datapack)
	logx.Info(b.GetChStatis().StringRev())

	return datapack, err
}

func (b *KcpChannel) IsReadLoopContinued(err error) bool {
	return false
}

func (b *KcpChannel) Write(datapack gch.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover error:", i)
			b.StopChannel(b)
		}
	}()

	if datapack.IsPrepare() {
		bytes := datapack.GetData()
		conn := b.conn
		conf := b.GetChConf()
		conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
		_, err := conn.Write(bytes)
		if err != nil {
			logx.Error("write error:", err)
			panic(err)
			return nil
		}
		gch.SendStatis(datapack)
		logx.Info(b.GetChStatis().StringSend())
		return err
	} else {
		logx.Info("packet is not prepare.")
	}
	return nil
}

type KcpPacket struct {
	gch.Basepacket
}

func (b *KcpChannel) NewPacket() gch.Packet {
	if b.protocol == gch.PROTOCOL_KWS00 {
		k := &KwsPacket{}
		k.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_KWS00)
		return k
	} else {
		k := &KcpPacket{}
		k.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_KCP)
		return k
	}
}

func (b *KcpChannel) StopChannel(channel gch.Channel) {
	if !b.IsClosed() {
		b.conn.Close()
	}
	b.BaseChannel.StopChannel(channel)
}

func (b *KcpChannel) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}

func (b *KcpChannel) RemoteAddr() net.Addr {
	return b.conn.RemoteAddr()
}

type KwsPacket struct {
	KcpPacket
	Frame Frame
}

func (packet *KwsPacket) SetData(data []byte) {
	packet.Basepacket.SetData(data)
	packet.Frame = NewInputFrame(data)
}
