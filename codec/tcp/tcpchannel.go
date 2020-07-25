/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcp

import (
	gchannel "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"net"
	"time"
)

type TcpChannel struct {
	gchannel.BaseChannel
	conn     *net.TCPConn
	readerr  *int
	writeerr *int
}

// TODO 配置化
var readbf []byte
func NewTcpChannel(conn *net.TCPConn, conf *config.ChannelConf) *TcpChannel {
	ch := &TcpChannel{conn: conn}
	ch.BaseChannel = *gchannel.NewBaseChannel(conf)
	bufSize := conf.ReadBufSize
	if bufSize <= 0 {
		bufSize = 10 * 1024
	}
	readbf = make([]byte, bufSize)
	return ch
}


func StartTcpChannel(conn *net.TCPConn, conf *config.ChannelConf, msgFunc gchannel.HandleMsgFunc) *TcpChannel {
	ch := NewTcpChannel(conn, conf)
	ch.SetHandleMsgFunc(msgFunc)
	go gchannel.StartReadLoop(ch)
	return ch
}

func (b *TcpChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *TcpChannel) Read() (packet gchannel.Packet, err error) {
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

func (b *TcpChannel) Write(datapack gchannel.Packet) error {
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

func (b *TcpChannel) NewPacket() gchannel.Packet {
	w := &TcpPacket{}
	w.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_TCP)
	return w
}

type TcpPacket struct {
	gchannel.Basepacket
}
