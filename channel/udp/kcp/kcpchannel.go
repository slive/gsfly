/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"errors"
	"fmt"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"gsfly/util"
	"net"
	"time"
)

type KcpChannel struct {
	gch.BaseChannel
	conn     *kcp.UDPSession
	protocol gch.Protocol
}

func newKcpChannel(kcpConn *kcp.UDPSession, chConf gch.ChannelConf, protocol gch.Protocol, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := &KcpChannel{conn: kcpConn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(chConf, chHandle)
	ch.protocol = protocol
	readBufSize := chConf.GetReadBufSize()
	kcpConn.SetReadBuffer(readBufSize)
	writeBufSize := chConf.GetWriteBufSize()
	kcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewSimpleKcpChannel(kcpConn *kcp.UDPSession, chConf gch.ChannelConf, msgFunc gch.OnMsgHandle) *KcpChannel {
	chHandle := gch.NewDefaultChHandle(msgFunc)
	return NewKcpChannel(kcpConn, chConf, chHandle)
}

func NewKcpChannel(kcpConn *kcp.UDPSession, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := newKcpChannel(kcpConn, chConf, chConf.GetProtocol(), chHandle)
	ch.SetChId("kcp-" + kcpConn.LocalAddr().String() + "-" + kcpConn.RemoteAddr().String() + "-" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

func (b *KcpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conn := b.conn
	conf := b.GetChConf()
	now := time.Now()
	// conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
		logx.Warn("read kcp err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil, nil
	}

	datapack := b.NewPacket()
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	logx.Info(b.GetChStatis().StringRev())

	return datapack, err
}

func (b *KcpChannel) IsReadLoopContinued(err error) bool {
	return false
}

func (b *KcpChannel) Write(datapack gch.Packet) error {
	if b.IsClosed() {
		return errors.New("kcpchannel had closed, chId:" + b.GetChId())
	}

	defer func() {
		rec := recover()
		if rec != nil {
			logx.Error("write kcp error, chId:%v, error:%v", b.GetChId(), rec)
			err, ok := rec.(error)
			if !ok {
				err = errors.New(fmt.Sprintf("%v", rec))
			}
			// 捕获处理消息异常
			b.GetChHandle().OnErrorHandle(b, util.NewError1(gch.ERR_WRITE, err))
			// 有异常，终止执行
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
			logx.Error("write kcp error:", err)
			gch.SendStatis(datapack, false)
			panic(err)
			return err
		}
		gch.SendStatis(datapack, true)
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
