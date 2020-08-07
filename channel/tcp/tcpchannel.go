/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcp

import (
	"errors"
	"fmt"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"gsfly/util"
	"net"
	"time"
)

type TcpChannel struct {
	gch.BaseChannel
	conn *net.TCPConn
}

func newTcpChannel(tcpConn *net.TCPConn, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *TcpChannel {
	ch := &TcpChannel{conn: tcpConn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(chConf, chHandle)
	readBufSize := chConf.GetReadBufSize()
	tcpConn.SetReadBuffer(readBufSize)

	writeBufSize := chConf.GetWriteBufSize()
	tcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewSimpleTcpChannel(tcpConn *net.TCPConn, chConf gch.ChannelConf, msgFunc gch.OnMsgHandle) *TcpChannel {
	chHandle := gch.NewDefaultChHandle(msgFunc)
	return NewTcpChannel(tcpConn, chConf, chHandle)
}

func NewTcpChannel(tcpConn *net.TCPConn, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *TcpChannel {
	ch := newTcpChannel(tcpConn, chConf, chHandle)
	ch.SetChId("tcp-" + tcpConn.LocalAddr().String() + "-" + tcpConn.RemoteAddr().String())
	return ch
}

func (b *TcpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conf := b.GetChConf()
	now := time.Now()
	b.conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		logx.Warn("read udp err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}

	bytes := readbf[0:readNum]
	datapack := b.NewPacket()
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	logx.Info(b.GetChStatis().StringRev())
	return datapack, err
}

func (b *TcpChannel) Write(datapack gch.Packet) error {
	if b.IsClosed() {
		return errors.New("tcpchannel had closed, chId:" + b.GetChId())
	}

	defer func() {
		rec := recover()
		if rec != nil {
			logx.Error("write tcp error, chId:%v, error:%v", b.GetChId(), rec)
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
		conf := b.GetChConf()
		b.conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
		_, err := b.conn.Write(bytes)
		if err != nil {
			logx.Error("write tcp error:", err)
			gch.SendStatis(datapack, false)
			panic(err)
			return err
		}
		gch.SendStatis(datapack, true)
		logx.Info(b.GetChStatis().StringSend())
		return err
	}
	return nil
}

func (b *TcpChannel) StopChannel(channel gch.Channel) {
	if !b.IsClosed() {
		b.conn.Close()
	}
	b.BaseChannel.StopChannel(channel)
}

func (b *TcpChannel) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}

func (b *TcpChannel) RemoteAddr() net.Addr {
	return b.conn.RemoteAddr()
}

func (b *TcpChannel) NewPacket() gch.Packet {
	w := &TcpPacket{}
	w.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_TCP)
	return w
}

// TcpPacket Tcp包
type TcpPacket struct {
	gch.Basepacket
}
