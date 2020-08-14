/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcpx

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type KcpChannel struct {
	gch.Channel
	Conn     *kcp.UDPSession
	protocol gch.Protocol
}

func NewKcpChannel(parent interface{}, kcpConn *kcp.UDPSession, chConf gch.IChannelConf, chHandle *gch.ChannelHandle) *KcpChannel {
	ch := &KcpChannel{Conn: kcpConn}
	ch.Channel = *gch.NewDefChannel(parent, chConf, chHandle)
	ch.protocol = chConf.GetProtocol()
	readBufSize := chConf.GetReadBufSize()
	kcpConn.SetReadBuffer(readBufSize)
	writeBufSize := chConf.GetWriteBufSize()
	kcpConn.SetWriteBuffer(writeBufSize)
	ch.SetId("kcp-" + kcpConn.LocalAddr().String() + "-" + kcpConn.RemoteAddr().String() + "-" + fmt.Sprintf("%v", kcpConn.GetConv()))
	return ch
}

func (b *KcpChannel) Read() (gch.IPacket, error) {
	return Read(b)
}

func Read(ch gch.IChannel) (gch.IPacket, error) {
	// TODO 超时配置
	conn := ch.GetConn()
	conf := ch.GetChConf()
	now := time.Now()
	// Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, err := conn.Read(readbf)
	if err != nil {
		// TODO 超时后抛出异常？
		logx.Warn("read kcp err:", err)
		gch.RevStatisFail(ch, now)
		return nil, err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil, nil
	}

	datapack := ch.NewPacket()
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	logx.Info(ch.GetChStatis().StringRev())

	return datapack, err
}

func (b *KcpChannel) IsReadLoopContinued(err error) bool {
	return false
}

func (b *KcpChannel) GetConn() net.Conn {
	return b.Conn
}

func (b *KcpChannel) Write(datapack gch.IPacket) error {
	return b.Channel.Write(datapack)
	// if b.IsClosed() {
	// 	return errors.New("kcpchannel had closed, chId:" + b.GetId())
	// }
	//
	// chHandle := b.GetChHandle()
	// defer func() {
	// 	rec := recover()
	// 	if rec != nil {
	// 		logx.Error("write kcp error, chId:%v, error:%v", b.GetId(), rec)
	// 		err, ok := rec.(error)
	// 		if !ok {
	// 			err = errors.New(fmt.Sprintf("%v", rec))
	// 		}
	// 		// 捕获处理消息异常
	// 		chHandle.OnErrorHandle(b, common.NewError1(gch.ERR_WRITE, err))
	// 		// 有异常，终止执行
	// 		b.StopChannel(b)
	// 	}
	// }()
	//
	// if datapack.IsPrepare() {
	// 	// 发送前的处理
	// 	befWriteHandle := chHandle.OnBefWriteHandle
	// 	if befWriteHandle != nil {
	// 		err := befWriteHandle(datapack)
	// 		if err != nil {
	// 			logx.Error("befWriteHandle error:", err)
	// 			return err
	// 		}
	// 	}
	//
	// 	bytes := datapack.GetData()
	// 	Conn := b.Conn
	// 	conf := b.GetChConf()
	// 	Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	// 	_, err := Conn.Write(bytes)
	// 	if err != nil {
	// 		logx.Error("write kcp error:", err)
	// 		gch.SendStatis(datapack, false)
	// 		panic(err)
	// 		return err
	// 	}
	// 	gch.SendStatis(datapack, true)
	// 	logx.Info(b.GetChStatis().StringSend())
	//
	// 	// 发送成功后的处理
	// 	aftWriteHandle := chHandle.OnAftWriteHandle
	// 	if aftWriteHandle != nil {
	// 		aftWriteHandle(datapack)
	// 	}
	// 	return err
	// } else {
	// 	logx.Warn("packet is not prepare.")
	// }
	// return nil
}

func (b *KcpChannel) WriteByConn(datapacket gch.IPacket) error {
	bytes := datapacket.GetData()
	conn := b.Conn
	conf := b.GetChConf()
	conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
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
	k.Packet = *gch.NewPacket(b, gch.PROTOCOL_KCP)
	return k
}

func (b *KcpChannel) Start() error {
	return b.StartChannel(b)
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