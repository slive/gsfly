/*
 * Author:slive
 * DATE:2020/7/17
 */
package udp

import (
	gchannel "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"net"
	"time"
)

type UdpChannel struct {
	gchannel.BaseChannel
	conn *net.UDPConn
}

// TODO 配置化
var readbf []byte
func NewUdpChannel(conn *net.UDPConn, conf *config.ChannelConf) *UdpChannel {
	ch := &UdpChannel{conn: conn}
	ch.BaseChannel = *gchannel.NewBaseChannel(conf)
	bufSize := conf.ReadBufSize
	if bufSize <= 0 {
		bufSize = 10 * 1024
	}
	readbf = make([]byte, bufSize)
	return ch
}

func StartUdpChannel(conn *net.UDPConn, conf *config.ChannelConf, msgFunc gchannel.HandleMsgFunc) *UdpChannel {
	ch := NewUdpChannel(conn, conf)
	ch.SetHandleMsgFunc(msgFunc)
	go gchannel.StartReadLoop(ch)
	return ch
}

func (b *UdpChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *UdpChannel) Read() (packet gchannel.Packet, err error) {
	// TODO 超时配置
	conf := b.GetConf()
	b.conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		return nil, err
	}

	bytes := readbf[0:readNum]
	datapack := b.NewPacket()
	datapack.SetData(bytes)
	logx.Info("receive:", string(bytes))
	return datapack, err
}

func (b *UdpChannel) Write(datapack gchannel.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover error:", i)
			b.Close()
		}
	}()
	if datapack.GetPrepare() {
		bytes := datapack.GetData()
		conf := b.GetConf()
		b.conn.SetWriteDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
		_, err := b.conn.Write(bytes)
		if err != nil {
			logx.Error("write error:", err)
			panic(err)
			return nil
		}
		logx.Info("write:", string(bytes))
		return err
	}
	return nil
}

func (b *UdpChannel) NewPacket() gchannel.Packet {
	w := &UdpPacket{}
	w.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_UDP)
	return w
}

type UdpPacket struct {
	gchannel.Basepacket
}
