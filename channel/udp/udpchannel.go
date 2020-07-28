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
	readBufSize := conf.ReadBufSize
	if readBufSize <= 0 {
		readBufSize = 10 * 1024
	}
	conn.SetReadBuffer(readBufSize)
	readbf = make([]byte, readBufSize)

	writeBufSize := conf.WriteBufSize
	if writeBufSize <= 0 {
		writeBufSize = 10 * 1024
	}
	conn.SetWriteBuffer(writeBufSize)
	return ch
}

func StartUdpChannel(wsconn *net.UDPConn, conf *config.ChannelConf, msgFunc gchannel.HandleMsgFunc) (*UdpChannel, error) {
	return StartUdpChannelWithHandle(wsconn, conf, gchannel.ChannelHandle{
		HandleMsgFunc: msgFunc,
	})
}

func StartUdpChannelWithHandle(conn *net.UDPConn, conf *config.ChannelConf, channelHandle gchannel.ChannelHandle) (*UdpChannel, error) {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start updchannel error:", re)
		}
	}()
	var err error
	ch := NewUdpChannel(conn, conf)
	ch.ChannelHandle = channelHandle
	go gchannel.StartReadLoop(ch)
	handle := ch.ChannelHandle
	startFunc := handle.HandleStartFunc
	if startFunc != nil {
		err = startFunc(ch)
	}
	return ch, err
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
	logx.Info("receive udp:", string(bytes))
	gchannel.RevStatis(datapack)
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
		logx.Info("write udp:", string(bytes))
		gchannel.SendStatis(datapack)
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
