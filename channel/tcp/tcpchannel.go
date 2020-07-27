/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcp

import (
	"gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"net"
	"time"
)

type TcpChannel struct {
	channel.BaseChannel
	conn     *net.TCPConn
	readerr  *int
	writeerr *int
}

// TODO 配置化
var readbf []byte

func NewTcpChannel(conn *net.TCPConn, conf *config.ChannelConf) *TcpChannel {
	ch := &TcpChannel{conn: conn}
	ch.BaseChannel = *channel.NewBaseChannel(conf)
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

func StartTcpChannel(conn *net.TCPConn, conf *config.ChannelConf, msgFunc channel.HandleMsgFunc) (*TcpChannel, error) {
	return StartTcpChannelWithHandle(conn, conf, channel.ChannelHandle{
		HandleMsgFunc: msgFunc,
	})
}

func StartTcpChannelWithHandle(conn *net.TCPConn, conf *config.ChannelConf, channelHandle channel.ChannelHandle) (*TcpChannel, error) {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start tcpchannel error:", re)
		}
	}()
	var err error
	ch := NewTcpChannel(conn, conf)
	ch.ChannelHandle = channelHandle
	go channel.StartReadLoop(ch)
	handle := ch.ChannelHandle
	startFunc := handle.HandleStartFunc
	if startFunc != nil {
		err = startFunc(ch)
	}
	return ch, err
}

func (b *TcpChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *TcpChannel) Read() (packet channel.Packet, err error) {
	// TODO 超时配置
	conf := config.Global_Conf.ChannelConf
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

func (b *TcpChannel) Write(datapack channel.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover error:", i)
			b.Close()
		}
	}()
	if datapack.GetPrepare() {
		bytes := datapack.GetData()
		conf := config.Global_Conf.ChannelConf
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

func (b *TcpChannel) NewPacket() channel.Packet {
	w := &TcpPacket{}
	w.Basepacket = *channel.NewBasePacket(b, channel.PROTOCOL_TCP)
	return w
}

type TcpPacket struct {
	channel.Basepacket
}
