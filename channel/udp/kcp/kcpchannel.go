/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	"gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"time"
)

type KcpChannel struct {
	channel.BaseChannel
	conn *kcp.UDPSession
}

type KChannel interface {
	channel.Channel
	GetConn() *kcp.UDPSession
}

// TODO 配置化
var readKcpBf []byte

func NewKcpChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf) *KcpChannel {
	ch := &KcpChannel{conn: kcpconn}
	ch.BaseChannel = *channel.NewBaseChannel(conf)
	readBufSize := conf.ReadBufSize
	if readBufSize <= 0 {
		readBufSize = 10 * 1024
	}
	kcpconn.SetReadBuffer(readBufSize)
	readKcpBf = make([]byte, readBufSize)

	writeBufSize := conf.WriteBufSize
	if writeBufSize <= 0 {
		writeBufSize = 10 * 1024
	}
	kcpconn.SetWriteBuffer(writeBufSize)
	return ch
}

func StartKcpChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf, msgFunc channel.HandleMsgFunc) (*KcpChannel, error) {
	return StartKcpChannelWithHandle(kcpconn, conf, channel.ChannelHandle{
		HandleMsgFunc: msgFunc,
	})
}

func StartKcpChannelWithHandle(kcpconn *kcp.UDPSession, conf *config.ChannelConf, channelHandle channel.ChannelHandle) (*KcpChannel, error) {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start kcpchannel error:", re)
		}
	}()
	var err error
	ch := NewKcpChannel(kcpconn, conf)
	ch.ChannelHandle = channelHandle
	go channel.StartReadLoop(ch)
	handle := ch.ChannelHandle
	startFunc := handle.HandleStartFunc
	if startFunc != nil {
		err = startFunc(ch)
	}
	return ch, err
}

func (b *KcpChannel) GetConn() *kcp.UDPSession {
	return b.conn
}

func (b *KcpChannel) GetChId() string {
	return b.conn.RemoteAddr().String() + ":" + fmt.Sprintf("%s", b.conn.GetConv())
}

func (b *KcpChannel) Read() (packet channel.Packet, err error) {
	// TODO 超时配置
	return ReadKcp(b)
}

func ReadKcp(b KChannel) (channel.Packet, error) {
	conf := b.GetConf()
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
	channel.RevStatis(datapack)
	return datapack, err
}

func (b *KcpChannel) Write(datapack channel.Packet) error {
	return WriteKcp(b, datapack)
}

func WriteKcp(b KChannel, datapack channel.Packet) error {
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
		conn := b.GetConn()
		conn.SetWriteDeadline(time.Now().Add(conf.WriteTimeout * time.Second))
		_, err := conn.Write(bytes)
		if err != nil {
			logx.Error("write error:", err)
			panic(err)
			return nil
		}
		logx.Info("write kcp:", string(bytes))
		channel.SendStatis(datapack)
		return err
	} else {
		logx.Info("packet is not prepare.")
	}
	return nil
}

type KcpPacket struct {
	channel.Basepacket
}

func (b *KcpChannel) NewPacket() channel.Packet {
	k := &KcpPacket{}
	k.Basepacket = *channel.NewBasePacket(b, channel.PROTOCOL_KCP)
	return k
}
