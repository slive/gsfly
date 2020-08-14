/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcpx

import (
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type TcpChannel struct {
	gch.Channel
	Conn *net.TCPConn
}

func newTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, chHandle *gch.ChannelHandle) *TcpChannel {
	ch := &TcpChannel{Conn: tcpConn}
	ch.Channel = *gch.NewDefChannel(parent, chConf, chHandle)
	readBufSize := chConf.GetReadBufSize()
	tcpConn.SetReadBuffer(readBufSize)

	writeBufSize := chConf.GetWriteBufSize()
	tcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewSimpleTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, msgFunc gch.OnMsgHandle) *TcpChannel {
	chHandle := gch.NewDefChHandle(msgFunc)
	return NewTcpChannel(parent, tcpConn, chConf, chHandle)
}

func NewTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, chHandle *gch.ChannelHandle) *TcpChannel {
	ch := newTcpChannel(parent, tcpConn, chConf, chHandle)
	ch.SetId("tcp-" + tcpConn.LocalAddr().String() + "-" + tcpConn.RemoteAddr().String())
	return ch
}

func (b *TcpChannel) Start() error {
	return b.StartChannel(b)
}

func (b *TcpChannel) Stop() {
	b.StopChannel(b)
}

func (b *TcpChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	conf := b.GetChConf()
	now := time.Now()
	b.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, err := b.Conn.Read(readbf)
	if err != nil {
		logx.Warn("read udp err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}

	bytes := readbf[0:readNum]
	datapack := b.NewPacket()
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	logx.Info(b.GetChStatis().StringRev())
	return datapack, err
}

func (b *TcpChannel) Write(datapack gch.IPacket) error {
	return b.Channel.Write(datapack)
	// if b.IsClosed() {
	// 	return errors.New("tcpchannel had closed, chId:" + b.GetId())
	// }
	//
	// chHandle := b.GetChHandle()
	// defer func() {
	// 	rec := recover()
	// 	if rec != nil {
	// 		logx.Error("write tcp error, chId:%v, error:%v", b.GetId(), rec)
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
	// if datapack.IsPrepare() {
	// 	// 发送前的处理
	// 	befWriteHandle := chHandle.OnBefWriteHandle
	// 	if befWriteHandle != nil {
	// 		err := befWriteHandle(datapack)
	// 		if err != nil{
	// 			logx.Error("befWriteHandle error:", err)
	// 			return err
	// 		}
	// 	}
	//
	// 	bytes := datapack.GetData()
	// 	conf := b.GetChConf()
	// 	b.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	// 	_, err := b.Conn.Write(bytes)
	// 	if err != nil {
	// 		logx.Error("write tcp error:", err)
	// 		gch.SendStatis(datapack, false)
	// 		panic(err)
	// 		return err
	// 	}
	//
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

func (b *TcpChannel) WriteByConn(datapacket gch.IPacket) error {
	bytes := datapacket.GetData()
	conf := b.GetChConf()
	b.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	_, err := b.Conn.Write(bytes)
	if err != nil {
		logx.Error("write tcp error:", err)
		gch.SendStatis(datapacket, false)
		panic(err)
		return err
	}
	return nil
}

func (b *TcpChannel) GetConn() net.Conn {
	return b.Conn
}

func (b *TcpChannel) LocalAddr() net.Addr {
	return b.Conn.LocalAddr()
}

func (b *TcpChannel) RemoteAddr() net.Addr {
	return b.Conn.RemoteAddr()
}

func (b *TcpChannel) NewPacket() gch.IPacket {
	w := &TcpPacket{}
	w.Packet = *gch.NewPacket(b, gch.PROTOCOL_TCP)
	return w
}

// TcpPacket Tcp包
type TcpPacket struct {
	gch.Packet
}
