/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	gconf "gsfly/config"
	logx "gsfly/logger"
	"time"
)

type KcpChannel struct {
	gch.BaseChannel
	conn *kcp.UDPSession
}

type KChannel interface {
	gch.Channel
	GetConn() *kcp.UDPSession
}

// TODO 配置化
var readKcpBf []byte

func newKcpChannel(kcpConn *kcp.UDPSession, conf *gconf.ChannelConf) *KcpChannel {
	ch := &KcpChannel{conn: kcpConn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	readBufSize := conf.ReadBufSize
	if readBufSize <= 0 {
		readBufSize = 10 * 1024
	}
	kcpConn.SetReadBuffer(readBufSize)
	readKcpBf = make([]byte, readBufSize)

	writeBufSize := conf.WriteBufSize
	if writeBufSize <= 0 {
		writeBufSize = 10 * 1024
	}
	kcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewKcpChannel(kcpConn *kcp.UDPSession, chConf *gconf.ChannelConf, msgFunc gch.HandleMsgFunc) *KcpChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewKcpChannelWithHandle(kcpConn, chConf, chHandle)
}

func NewKcpChannelWithHandle(kcpConn *kcp.UDPSession, chConf *gconf.ChannelConf, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := newKcpChannel(kcpConn, chConf)
	ch.ChannelHandle = *chHandle
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

func (b *KcpChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String() + ":" + fmt.Sprintf("%v", b.conn.GetConv())
}

func (b *KcpChannel) Read() (packet gch.Packet, err error) {
	// TODO 超时配置
	return ReadKcp(b)
}

func ReadKcp(b KChannel) (gch.Packet, error) {
	conf := b.GetChConf()
	conn := b.GetConn()
	conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	readNum, err := conn.Read(readKcpBf)
	if err != nil {
		return nil, err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil, nil
	}

	datapack := b.NewPacket()
	bytes := readKcpBf[0:readNum]
	datapack.SetData(bytes)
	logx.Info("receive kcp:" + string(bytes))
	gch.RevStatis(datapack)
	return datapack, err
}

func (b *KcpChannel) Write(datapack gch.Packet) error {
	return WriteKcp(b, datapack)
}

func WriteKcp(b KChannel, datapack gch.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover error:", i)
			b.StopChannel()
		}
	}()

	if datapack.IsPrepare() {
		bytes := datapack.GetData()
		conf := b.GetChConf()
		conn := b.GetConn()
		conn.SetWriteDeadline(time.Now().Add(conf.WriteTimeout * time.Second))
		_, err := conn.Write(bytes)
		if err != nil {
			logx.Error("write error:", err)
			panic(err)
			return nil
		}
		logx.Info("write kcp:", string(bytes))
		gch.SendStatis(datapack)
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
	k := &KcpPacket{}
	k.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_KCP)
	return k
}
