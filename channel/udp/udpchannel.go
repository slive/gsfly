/*
 * Udp通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package udp

import (
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type UdpChannel struct {
	gch.BaseChannel
	conn *net.UDPConn
}

func newUdpChannel(conn *net.UDPConn, conf *gch.BaseChannelConf) *UdpChannel {
	ch := &UdpChannel{conn: conn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	readBufSize := conf.GetReadBufSize()
	conn.SetReadBuffer(readBufSize)

	writeBufSize := conf.GetWriteBufSize()
	conn.SetWriteBuffer(writeBufSize)
	return ch
}

// NewUdpChannel 创建udpchannel，需实现handleMsgFunc方法
func NewUdpChannel(udpConn *net.UDPConn, chConf *gch.BaseChannelConf, msgFunc gch.HandleMsgFunc) *UdpChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewUdpChannelWithHandle(udpConn, chConf, chHandle)
}

// NewUdpChannelWithHandle 创建udpchannel，需实现ChannelHandle
func NewUdpChannelWithHandle(udpConn *net.UDPConn, chConf *gch.BaseChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := newUdpChannel(udpConn, chConf)
	ch.ChannelHandle = *chHandle
	ch.SetChId("udp-" + udpConn.LocalAddr().String() + "-" + udpConn.RemoteAddr().String())
	return ch
}

func (b *UdpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conf := b.GetChConf()
	b.conn.SetReadDeadline(time.Now().Add(conf.GetReadTimeout() * time.Second))
	// TODO 是否有性能问题？
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, addr, err := b.conn.ReadFrom(readbf)
	if err != nil {
		return nil, err
	}

	bytes := readbf[0:readNum]
	nPacket := b.NewPacket()
	datapack := nPacket.(*UdpPacket)
	datapack.SetData(bytes)
	datapack.Addr = addr
	logx.Info(b.GetChStatis().StringRev())
	gch.RevStatis(datapack)
	return datapack, err
}

// Write datapack需要设置目标addr
func (b *UdpChannel) Write(datapack gch.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("write error:", i)
			// 有异常，终止执行
			b.StopChannel(b)
		}
	}()
	if datapack.IsPrepare() {
		writePacket := datapack.(*UdpPacket)
		bytes := writePacket.GetData()
		conf := b.GetChConf()
		b.conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
		addr := writePacket.Addr
		var err error
		if addr != nil {
			_, err = b.conn.WriteTo(bytes, addr)
		} else {
			_, err = b.conn.Write(bytes)
		}
		if err != nil {
			logx.Error("write error:", err)
			panic(err)
			return err
		}
		gch.SendStatis(datapack)
		logx.Info(b.GetChStatis().StringSend())
		return err
	}
	return nil
}

func (b *UdpChannel) StopChannel(channel gch.Channel) {
	if !b.IsClosed() {
		b.conn.Close()
	}
	b.BaseChannel.StopChannel(channel)
}

func (b *UdpChannel) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}

func (b *UdpChannel) RemoteAddr() net.Addr {
	return b.conn.RemoteAddr()
}

func (b *UdpChannel) NewPacket() gch.Packet {
	w := &UdpPacket{}
	w.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_UDP)
	return w
}

// UdpPacket Udp包
type UdpPacket struct {
	gch.Basepacket
	Addr net.Addr
}
