/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcpx

import (
	"fmt"
	gch "github.com/slive/gsfly/channel"
	logx "github.com/slive/gsfly/logger"
	"github.com/xtaci/kcp-go"
	"net"
	"time"
)

// KcpChannel
type KcpChannel struct {
	gch.Channel
	Conn             *kcp.UDPSession
	protocol         gch.Network
	onKcpReadHandler gch.ChHandleFunc
	connected        bool
}

// NewKcpChannel 创建KcpChannel
func NewKcpChannel(parent interface{}, kcpConn *kcp.UDPSession, chConf gch.IChannelConf, chHandle *gch.ChHandle, server bool) *KcpChannel {
	ch := &KcpChannel{Conn: kcpConn}
	ch.Channel = *gch.NewDefChannel(parent, chConf, chHandle, server)
	ch.protocol = chConf.GetNetwork()
	readBufSize := chConf.GetReadBufSize()
	kcpConn.SetReadBuffer(readBufSize)
	writeBufSize := chConf.GetWriteBufSize()
	kcpConn.SetWriteBuffer(writeBufSize)
	// TODO kcp相关配置，待具体实现
	kcpConn.SetACKNoDelay(true)
	kcpConn.SetWriteDelay(false)
	ch.onKcpReadHandler = chHandle.GetOnRead()

	// 重新包装
	chHandle.SetOnRead(ch.onKcpRead)
	ch.SetId(kcpConn.LocalAddr().String() + "->" + kcpConn.RemoteAddr().String() + "#" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

func (b *KcpChannel) Read() (gch.IPacket, error) {
	return Read(b)
}

// onKcpRead kcp的读取处理
func (b *KcpChannel) onKcpRead(ctx gch.IChHandleContext) {
	handleFunc := b.onKcpReadHandler
	if handleFunc != nil {
		var continued = true
		if !b.connected {
			// 第一次使用时为链接
			gch.HandleOnConnnect(ctx)
			continued = (ctx.GetError() == nil)
			b.connected = continued
		}

		if continued {
			handleFunc(ctx)
		} else {
			logx.Error("onKcpRead error:%v.", ctx.GetError())
		}
	}
}

func Read(ch *KcpChannel) (gch.IPacket, error) {
	// TODO 超时配置
	conn := ch.GetConn()
	conf := ch.GetConf()
	now := time.Now()
	conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := ch.NewReadBuf()
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
		logx.Warn("read kcp err:", err)
		gch.RevStatisFail(ch, now)
		return nil, err
	}
	// TODO 接收到8个字节数据，是否要忽略
	// if readNum <= 8 {
	// 	return nil, nil
	// }

	datapack := ch.NewPacket()
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	return datapack, err
}

func (b *KcpChannel) GetConn() net.Conn {
	return b.Conn
}

func (b *KcpChannel) Write(datapack gch.IPacket) error {
	return b.Channel.Write(datapack)
}

func (b *KcpChannel) WriteByConn(datapacket gch.IPacket) error {
	bytes := datapacket.GetData()
	conn := b.Conn
	conf := b.GetConf()
	writeDeadLine := time.Now().Add(conf.GetWriteTimeout() * time.Second)
	logx.Info("writeDeadLine:", writeDeadLine)
	conn.SetWriteDeadline(writeDeadLine)
	_, err := conn.Write(bytes)
	if err != nil {
		logx.Error("write kcp error:", err)
		gch.SendStatis(datapacket, false)
		panic(err)
		return err
	}
	return nil
}

type KcpPacket struct {
	gch.Packet
}

func (b *KcpChannel) NewPacket() gch.IPacket {
	k := &KcpPacket{}
	k.Packet = *gch.NewPacket(b, gch.NETWORK_KCP)
	return k
}

func (b *KcpChannel) Open() error {
	return b.StartChannel(b)
}

func (b *KcpChannel) Release() {
	b.connected = false
	b.StopChannel(b)
}

func (b *KcpChannel) LocalAddr() net.Addr {
	return b.Conn.LocalAddr()
}

func (b *KcpChannel) RemoteAddr() net.Addr {
	return b.Conn.RemoteAddr()
}
