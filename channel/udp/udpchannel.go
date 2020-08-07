/*
 * Udp通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package udp

import (
	"fmt"
	"github.com/pkg/errors"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"gsfly/util"
	"net"
	"time"
)

type UdpChannel struct {
	gch.BaseChannel
	conn *net.UDPConn
}

func newUdpChannel(conn *net.UDPConn, conf gch.ChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := &UdpChannel{conn: conn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf, chHandle)
	readBufSize := conf.GetReadBufSize()
	conn.SetReadBuffer(readBufSize)

	writeBufSize := conf.GetWriteBufSize()
	conn.SetWriteBuffer(writeBufSize)
	return ch
}

// NewSimpleUdpChannel 创建udpchannel，需实现handleMsgFunc方法
func NewSimpleUdpChannel(udpConn *net.UDPConn, chConf gch.ChannelConf, msgFunc gch.OnMsgHandle) *UdpChannel {
	chHandle := gch.NewDefaultChHandle(msgFunc)
	return NewUdpChannel(udpConn, chConf, chHandle)
}

// NewUdpChannel 创建udpchannel，需实现ChannelHandle
func NewUdpChannel(udpConn *net.UDPConn, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := newUdpChannel(udpConn, chConf, chHandle)
	ch.SetChId("udp-" + udpConn.LocalAddr().String() + "-" + udpConn.RemoteAddr().String())
	return ch
}

func (b *UdpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conf := b.GetChConf()
	now := time.Now()
	b.conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	// TODO 是否有性能问题？
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, addr, err := b.conn.ReadFrom(readbf)
	if err != nil {
		logx.Warn("read udp err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}

	bytes := readbf[0:readNum]
	nPacket := b.NewPacket()
	datapack := nPacket.(*UdpPacket)
	datapack.SetData(bytes)
	datapack.Addr = addr
	logx.Info(b.GetChStatis().StringRev())
	gch.RevStatis(datapack, true)
	return datapack, err
}

// Write datapack需要设置目标addr
func (b *UdpChannel) Write(datapack gch.Packet) error {
	if b.IsClosed() {
		return errors.New("udpchannel had closed, chId:" + b.GetChId())
	}

	defer func() {
		rec := recover()
		if rec != nil {
			logx.Error("write udp error, chId:%v, error:%v", b.GetChId(), rec)
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
			logx.Error("write udp error:", err)
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
