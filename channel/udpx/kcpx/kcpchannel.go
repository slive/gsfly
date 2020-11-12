/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcpx

import (
	"fmt"
	gch "github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
	"github.com/xtaci/kcp-go"
	"io"
	"net"
	"time"
)

// KcpChannel
type KcpChannel struct {
	gch.Channel
	Conn             *kcp.UDPSession
	protocol         gch.Network
	onKcpReadHandler gch.ChHandleFunc
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
	kcpConn.SetACKNoDelay(true)
	kcpConn.SetWriteDelay(false)
	ch.onKcpReadHandler = chHandle.GetOnRead()
	// 重新包装
	chHandle.SetOnRead(ch.onKcpWapperReadHandler)
	ch.SetId("kcp-" + kcpConn.LocalAddr().String() + "-" + kcpConn.RemoteAddr().String() + "-" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

func (b *KcpChannel) Read() (gch.IPacket, error) {
	return Read(b)
}

func (b *KcpChannel) onKcpWapperReadHandler(ctx gch.IChHandleContext) {
	handleFunc := b.onKcpReadHandler
	if handleFunc != nil {
		if b.IsServer() {
			// 激活处理方法
			if !b.IsActived() {
				gch.HandleOnActive(ctx)
				b.SetActived(true)
			}
			packet := ctx.GetPacket()
			if packet != nil && !packet.IsRelease() {
				handleFunc(ctx)
			}
		} else {
			handleFunc(ctx)
		}
	}
}

func Read(ch *KcpChannel) (gch.IPacket, error) {
	// TODO 超时配置
	conn := ch.GetConn()
	conf := ch.GetConf()
	now := time.Now()
	conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := ch.GetReadBuf()
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
		logx.Warn("read kcp err:", err)
		gch.RevStatisFail(ch, now)
		if err == io.EOF {
			panic(err)
		}
		return nil, err
	}
	// 接收到8个字节数据，是bug?
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

func (b *KcpChannel) Start() error {
	err := b.StartChannel(b)
	if err == nil && !b.IsServer() {
		// 客户端启动即激活？
		gch.HandleOnActive(gch.NewChHandleContext(b, nil))
	}
	return err
}

func (b *KcpChannel) Stop() {
	b.StopChannel(b)
}

func (b *KcpChannel) LocalAddr() net.Addr {
	return b.Conn.LocalAddr()
}

func (b *KcpChannel) RemoteAddr() net.Addr {
	return b.Conn.RemoteAddr()
}
