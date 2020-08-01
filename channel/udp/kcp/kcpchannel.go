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

func newKcpChannel(kcpConn *kcp.UDPSession, conf *gch.ChannelConf, protocol gch.Protocol) *KcpChannel {
	ch := &KcpChannel{conn: kcpConn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	ch.protocol = protocol
	readBufSize := conf.ReadBufSize
	if readBufSize <= 0 {
		readBufSize = 10 * 1024
	}

	writeBufSize := conf.WriteBufSize
	if writeBufSize <= 0 {
		writeBufSize = 10 * 1024
	}
	kcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewKcpChannel(kcpConn *kcp.UDPSession, chConf *gch.ChannelConf, msgFunc gch.HandleMsgFunc) *KcpChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewKcpChannelWithHandle(kcpConn, chConf, chHandle)
}

func NewKcpChannelWithHandle(kcpConn *kcp.UDPSession, chConf *gch.ChannelConf, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := newKcpChannel(kcpConn, chConf, chConf.Protocol)
	ch.ChannelHandle = *chHandle
	ch.SetChId(kcpConn.LocalAddr().String() + ":" + kcpConn.RemoteAddr().String() + ":" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

// func (b *KcpChannel) StartChannel() error {
// 	defer func() {
// 		re := recover()
// 		if re != nil {
// 			logx.Error("Start kcpChannel error:", re)
// 		}
// 	}()
// 	go b.StartReadLoop(b)
//
// 	// 启动后处理方法
// 	startFunc := b.HandleStartFunc
// 	if startFunc != nil {
// 		err := startFunc(b)
// 		if err != nil {
// 			b.StopChannel()
// 		}
// 		return err
// 	}
// 	return nil
// }

func (b *KcpChannel) GetConn() *kcp.UDPSession {
	return b.conn
}

func (b *KcpChannel) Read() (packet gch.Packet, err error) {
	// TODO 超时配置
	conn := b.GetConn()
	conf := b.GetChConf()
	conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	readbf := make([]byte, conf.ReadBufSize)
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
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
		conn := b.GetConn()
		conf := b.GetChConf()
		conn.SetWriteDeadline(time.Now().Add(conf.WriteTimeout * time.Second))
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
