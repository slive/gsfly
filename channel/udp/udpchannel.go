/*
 * Udp通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package udp

import (
	gch "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"net"
	"time"
)

type UdpChannel struct {
	gch.BaseChannel
	conn   *net.UDPConn
	readbf []byte
}

func newUdpChannel(conn *net.UDPConn, conf *config.ChannelConf) *UdpChannel {
	ch := &UdpChannel{conn: conn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	readBufSize := conf.ReadBufSize
	if readBufSize <= 0 {
		readBufSize = 10 * 1024
	}
	conn.SetReadBuffer(readBufSize)
	ch.readbf = make([]byte, readBufSize)

	writeBufSize := conf.WriteBufSize
	if writeBufSize <= 0 {
		writeBufSize = 10 * 1024
	}
	conn.SetWriteBuffer(writeBufSize)
	return ch
}

// NewUdpChannel 创建udpchannel，需实现handleMsgFunc方法
func NewUdpChannel(udpConn *net.UDPConn, chConf *config.ChannelConf, msgFunc gch.HandleMsgFunc) *UdpChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewUdpChannelWithHandle(udpConn, chConf, chHandle)
}

// NewUdpChannelWithHandle 创建udpchannel，需实现ChannelHandle
func NewUdpChannelWithHandle(udpConn *net.UDPConn, chConf *config.ChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := newUdpChannel(udpConn, chConf)
	ch.ChannelHandle = *chHandle
	ch.SetChId(udpConn.LocalAddr().String() + ":" + udpConn.RemoteAddr().String())
	return ch
}

func (b *UdpChannel) Read() (packet gch.Packet, err error) {
	// TODO 超时配置
	conf := b.GetChConf()
	b.conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	readbf := b.readbf
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		return nil, err
	}

	bytes := readbf[0:readNum]
	datapack := b.NewPacket()
	datapack.SetData(bytes)
	logx.Info(b.GetChStatis().StringRev())
	gch.RevStatis(datapack)
	return datapack, err
}

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
		bytes := datapack.GetData()
		conf := b.GetChConf()
		b.conn.SetWriteDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
		_, err := b.conn.Write(bytes)
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
}
