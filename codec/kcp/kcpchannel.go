/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	gchannel "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"time"
)

type KcpChannel struct {
	gchannel.BaseChannel
	conn *kcp.UDPSession
}

type Channel interface {
	gchannel.Channel
	GetConn() *kcp.UDPSession
}

// TODO 配置化
var readbf []byte

func NewKcpChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf) *KcpChannel {
	ch := &KcpChannel{conn: kcpconn}
	ch.BaseChannel = *gchannel.NewBaseChannel(conf)
	bufSize := conf.ReadBufSize
	if bufSize <= 0 {
		bufSize = 10 * 1024
	}
	readbf = make([]byte, bufSize)
	return ch
}

func StartKcpChannel(kcpconn *kcp.UDPSession, conf *config.ChannelConf, msgFunc gchannel.HandleMsgFunc) *KcpChannel {
	ch := NewKcpChannel(kcpconn, conf)
	ch.SetHandleMsgFunc(handlerMessage)
	go gchannel.StartReadLoop(ch)
	return ch
}

func (b *KcpChannel) GetConn() *kcp.UDPSession {
	return b.conn
}

func (b *KcpChannel) GetChId() string {
	return b.conn.RemoteAddr().String() + ":" + fmt.Sprintf("%s", b.conn.GetConv())
}

func (b *KcpChannel) Read() (packet gchannel.Packet, err error) {
	// TODO 超时配置
	return ReadKcp(b)
}

func ReadKcp(b Channel) (gchannel.Packet, error) {
	conf := b.GetConf()
	conn := b.GetConn()
	conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	readNum, err := conn.Read(readbf)
	if err != nil {
		return nil, err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil, nil
	}

	datapack := b.NewPacket()
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	logx.Info("receive:" + string(bytes))
	return datapack, err
}

func (b *KcpChannel) Write(datapack gchannel.Packet) error {
	return WriteKcp(b, datapack)
}

func WriteKcp(b Channel, datapack gchannel.Packet) error {
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
		conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
		_, err := conn.Write(bytes)
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

func handlerMessage(datapack gchannel.Packet) error {
	logx.Info("handler datapack:", datapack)
	return nil
}

type KcpPacket struct {
	gchannel.Basepacket
}

func (b *KcpChannel) NewPacket() gchannel.Packet {
	k := &KcpPacket{}
	k.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_KCP)
	return k
}
