/*
 * Udp通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package udpx

import (
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type UdpChannel struct {
	gch.BaseChannel
	Conn *net.UDPConn
}

func newUdpChannel(parent interface{}, conn *net.UDPConn, conf gch.ChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := &UdpChannel{Conn: conn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(parent, conf, chHandle)
	readBufSize := conf.GetReadBufSize()
	conn.SetReadBuffer(readBufSize)

	writeBufSize := conf.GetWriteBufSize()
	conn.SetWriteBuffer(writeBufSize)
	return ch
}

// NewSimpleUdpChannel 创建udpchannel，需实现handleMsgFunc方法
func NewSimpleUdpChannel(parent interface{}, udpConn *net.UDPConn, chConf gch.ChannelConf, msgFunc gch.OnMsgHandle) *UdpChannel {
	chHandle := gch.NewDefaultChHandle(msgFunc)
	return NewUdpChannel(parent, udpConn, chConf, chHandle)
}

// NewUdpChannel 创建udpchannel，需实现ChannelHandle
func NewUdpChannel(parent interface{}, udpConn *net.UDPConn, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *UdpChannel {
	ch := newUdpChannel(parent, udpConn, chConf, chHandle)
	ch.SetId("udp-" + udpConn.LocalAddr().String() + "-" + udpConn.RemoteAddr().String())
	return ch
}

func (b *UdpChannel) Start() error {
	return b.StartChannel(b)
}

func (b *UdpChannel) Stop() {
	b.StopChannel(b)
}

func (b *UdpChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	conf := b.GetChConf()
	now := time.Now()
	b.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	// TODO 是否有性能问题？
	readbf := make([]byte, conf.GetReadBufSize())
	readNum, addr, err := b.Conn.ReadFrom(readbf)
	if err != nil {
		logx.Warn("read udp err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}

	bytes := readbf[0:readNum]
	nPacket := b.NewPacket()
	datapack := nPacket.(*UdpPacket)
	datapack.SetData(bytes)
	datapack.Addr = addr
	logx.Info(b.GetChStatis().StringRev())
	gch.RevStatis(datapack, true)
	return datapack, err
}

// Write datapack需要设置目标addr
func (b *UdpChannel) Write(datapack gch.Packet) error {
	return b.BaseChannel.Write(datapack)
	// 	if b.IsClosed() {
	// 		return errors.New("udpchannel had closed, chId:" + b.GetId())
	// 	}
	//
	// 	chHandle := b.GetChHandle()
	// 	defer func() {
	// 		rec := recover()
	// 		if rec != nil {
	// 			logx.Error("write udp error, chId:%v, error:%v", b.GetId(), rec)
	// 			err, ok := rec.(error)
	// 			if !ok {
	// 				err = errors.New(fmt.Sprintf("%v", rec))
	// 			}
	// 			// 捕获处理消息异常
	// 			chHandle.OnErrorHandle(b, common.NewError1(gch.ERR_WRITE, err))
	// 			// 有异常，终止执行
	// 			b.StopChannel(b)
	// 		}
	// 	}()
	// 	if datapack.IsPrepare() {
	// 		// 发送前的处理
	// 		befWriteHandle := chHandle.OnBefWriteHandle
	// 		if befWriteHandle != nil {
	// 			err := befWriteHandle(datapack)
	// 			if err != nil {
	// 				logx.Error("befWriteHandle error:", err)
	// 				return err
	// 			}
	// 		}
	//
	// 		gch.SendStatis(datapack, true)
	// 		logx.Info(b.GetChStatis().StringSend())
	//
	// 		// 发送成功后的处理
	// 		aftWriteHandle := chHandle.OnAftWriteHandle
	// 		if aftWriteHandle != nil {
	// 			aftWriteHandle(writePacket)
	// 		}
	// 		return err
	// 	} else {
	// 		logx.Warn("packet is not prepare.")
	// 	}
	// 	return nil
}

// WriteByConn 实现通过conn发送
func (b *UdpChannel) WriteByConn(datapacket gch.Packet) error {
	writePacket := datapacket.(*UdpPacket)
	bytes := writePacket.GetData()
	conf := b.GetChConf()
	// 设置超时时间
	b.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	addr := writePacket.Addr
	var err error
	if addr != nil {
		// 发送到某个目标地址
		_, err = b.Conn.WriteTo(bytes, addr)
	} else {
		// 无目标地址发送
		_, err = b.Conn.Write(bytes)
	}
	if err != nil {
		logx.Error("write udp error:", err)
		gch.SendStatis(datapacket, false)
		panic(err)
		return err
	}
	return nil
}

func (b *UdpChannel) LocalAddr() net.Addr {
	return b.Conn.LocalAddr()
}

func (b *UdpChannel) RemoteAddr() net.Addr {
	return b.Conn.RemoteAddr()
}

func (b *UdpChannel) GetConn() net.Conn {
	return b.Conn
}

func (b *UdpChannel) NewPacket() gch.Packet {
	w := &UdpPacket{}
	w.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_UDP)
	return w
}

// UdpPacket Udp包
type UdpPacket struct {
	gch.Basepacket
	Addr net.Addr
}
