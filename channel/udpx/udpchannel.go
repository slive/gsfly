/*
 * Udp通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package udpx

import (
	gch "github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
	"github.com/pkg/errors"
	"net"
	"time"
)

type UdpChannel struct {
	gch.Channel
	Conn     *net.UDPConn
	rAddr    *net.UDPAddr
	readchan chan []byte
}
// udp 包最大不大于65535
const Max_UDP_Buf = 65535

func newUdpChannel(parent interface{}, conn *net.UDPConn, conf gch.IChannelConf, chHandle *gch.ChHandle, rAddr *net.UDPAddr, server bool) *UdpChannel {
	ch := &UdpChannel{Conn: conn}
	ch.Channel = *gch.NewDefChannel(parent, conf, chHandle, server)
	readBufSize := conf.GetReadBufSize()
	if readBufSize > Max_UDP_Buf {
		readBufSize = Max_UDP_Buf
	}
	conn.SetReadBuffer(readBufSize)

	writeBufSize := conf.GetWriteBufSize()
	if writeBufSize > Max_UDP_Buf {
		writeBufSize = Max_UDP_Buf
	}
	conn.SetWriteBuffer(writeBufSize)
	ch.rAddr = rAddr
	if server {
		ch.readchan = make(chan []byte, 100)
	}
	return ch
}

// NewSimpleUdpChannel 创建udpchannel，需实现handleMsgFunc方法
func NewSimpleUdpChannel(parent interface{}, udpConn *net.UDPConn, chConf gch.IChannelConf, onReadHandler gch.ChHandleFunc, rAddr *net.UDPAddr, server bool) *UdpChannel {
	chHandle := gch.NewDefChHandle(onReadHandler)
	return NewUdpChannel(parent, udpConn, chConf, chHandle, rAddr, server)
}

// NewUdpChannel 创建udpchannel，需实现ChannelHandle
func NewUdpChannel(parent interface{}, udpConn *net.UDPConn, chConf gch.IChannelConf, chHandle *gch.ChHandle, rAddr *net.UDPAddr, server bool) *UdpChannel {
	ch := newUdpChannel(parent, udpConn, chConf, chHandle, rAddr, server)
	ch.SetId(FetchUdpId(udpConn, rAddr))
	return ch
}

func FetchUdpId(udpConn *net.UDPConn, rAddr *net.UDPAddr) string {
	return udpConn.LocalAddr().String() + "->" + rAddr.String()
}

func (udpCh *UdpChannel) Open() error {
	err := udpCh.StartChannel(udpCh)
	if err == nil {
		gch.HandleOnActive(gch.NewChHandleContext(udpCh, nil))
	}
	return err
}

func (udpCh *UdpChannel) Close() {
	udpCh.StopChannel(udpCh)
	if udpCh.IsServer() {
		close(udpCh.readchan)
	}
}

func (udpCh *UdpChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	raddr := udpCh.rAddr
	var bytes []byte
	if !udpCh.IsServer() {
		readbf := udpCh.NewReadBuf()
		conf := udpCh.GetConf()
		now := time.Now()
		udpCh.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
		// TODO 是否有性能问题？
		readNum, addr, err := udpCh.Conn.ReadFromUDP(readbf)
		if err != nil {
			logx.Warn("read udp err:", err)
			gch.RevStatisFail(udpCh, now)
			return nil, err
		}
		raddr = addr
		bytes = readbf[0:readNum]
	} else {
		// 服务端直接获取
		bytes = <-udpCh.readchan
	}

	if len(bytes) > 0 {
		nPacket := udpCh.NewPacket()
		datapack := nPacket.(*UdpPacket)
		datapack.SetData(bytes)
		datapack.RAddr = raddr
		gch.RevStatis(datapack, true)
		return datapack, nil
	}
	return nil, errors.New("udp data is null.")
}

func (udpCh *UdpChannel) CacheServerRead(bytes []byte) {
	udpCh.readchan <- bytes
}

// Write datapack需要设置目标addr
func (udpCh *UdpChannel) Write(datapack gch.IPacket) error {
	return udpCh.Channel.Write(datapack)
}

// WriteByConn 实现通过conn发送
func (udpCh *UdpChannel) WriteByConn(datapacket gch.IPacket) error {
	writePacket := datapacket.(*UdpPacket)
	bytes := writePacket.GetData()
	conf := udpCh.GetConf()
	// 设置超时时间
	udpCh.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	addr := writePacket.RAddr
	var err error
	if addr != nil {
		// 发送到某个目标地址
		_, err = udpCh.Conn.WriteTo(bytes, addr)
	} else {
		// 无目标地址发送
		_, err = udpCh.Conn.Write(bytes)
	}
	if err != nil {
		logx.Error("write udp error:", err)
		gch.SendStatis(datapacket, false)
		panic(err)
		return err
	}
	return nil
}

func (udpCh *UdpChannel) LocalAddr() net.Addr {
	return udpCh.Conn.LocalAddr()
}

func (udpCh *UdpChannel) RemoteAddr() net.Addr {
	return udpCh.Conn.RemoteAddr()
}

func (udpCh *UdpChannel) GetConn() net.Conn {
	return udpCh.Conn
}

func (udpCh *UdpChannel) NewPacket() gch.IPacket {
	w := &UdpPacket{}
	w.Packet = *gch.NewPacket(udpCh, gch.NETWORK_UDP)
	w.RAddr = udpCh.rAddr
	return w
}

// UdpPacket Udp包
type UdpPacket struct {
	gch.Packet
	RAddr *net.UDPAddr
}
