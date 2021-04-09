/*
 * 基于channel实现的tcpchannel功能，因tcp是协议层，需用户根据各自需要定制上层协议。
 *
 * Author:slive
 * DATE:2020/7/17
 */
package tcpx

import (
	gch "github.com/slive/gsfly/channel"
	logx "github.com/slive/gsfly/logger"
	"net"
	"time"
)

type TcpChannel struct {
	gch.Channel
	Conn *net.TCPConn
}

func newTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, chHandle *gch.ChHandle, server bool) *TcpChannel {
	ch := &TcpChannel{Conn: tcpConn}
	ch.Channel = *gch.NewDefChannel(parent, chConf, chHandle, server)
	readBufSize := chConf.GetReadBufSize()
	tcpConn.SetReadBuffer(readBufSize)

	writeBufSize := chConf.GetWriteBufSize()
	tcpConn.SetWriteBuffer(writeBufSize)
	return ch
}

func NewSimpleTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, onReadHandler gch.ChHandleFunc, server bool) *TcpChannel {
	chHandle := gch.NewDefChHandle(onReadHandler)
	return NewTcpChannel(parent, tcpConn, chConf, chHandle, server)
}

func NewTcpChannel(parent interface{}, tcpConn *net.TCPConn, chConf gch.IChannelConf, chHandle *gch.ChHandle, server bool) *TcpChannel {
	ch := newTcpChannel(parent, tcpConn, chConf, chHandle, server)
	ch.SetId(tcpConn.LocalAddr().String() + "->" + tcpConn.RemoteAddr().String())
	return ch
}

func (tcpCh *TcpChannel) Open() error {
	err := tcpCh.StartChannel(tcpCh)
	if err == nil {
		gch.HandleOnConnnect(gch.NewChHandleContext(tcpCh, nil))
	}
	return err
}

func (tcpCh *TcpChannel) Release() {
	tcpCh.StopChannel(tcpCh)
}

func (tcpCh *TcpChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	conf := tcpCh.GetConf()
	now := time.Now()
	tcpCh.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	readbf := tcpCh.NewReadBuf()
	readNum, err := tcpCh.Conn.Read(readbf)
	if err != nil {
		logx.Warn("read udp err:", err)
		gch.RevStatisFail(tcpCh, now)
		return nil, err
	}

	bytes := readbf[0:readNum]
	datapack := tcpCh.NewPacket()
	datapack.SetData(bytes)
	gch.RevStatis(datapack, true)
	return datapack, err
}

func (tcpCh *TcpChannel) Write(datapack gch.IPacket) error {
	return tcpCh.Channel.Write(datapack)
}

func (tcpCh *TcpChannel) WriteByConn(datapacket gch.IPacket) error {
	bytes := datapacket.GetData()
	conf := tcpCh.GetConf()
	tcpCh.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	_, err := tcpCh.Conn.Write(bytes)
	if err != nil {
		logx.Error("write tcp error:", err)
		gch.SendStatis(datapacket, false)
		panic(err)
		return err
	}
	return nil
}

func (tcpCh *TcpChannel) GetConn() net.Conn {
	return tcpCh.Conn
}

func (tcpCh *TcpChannel) LocalAddr() net.Addr {
	return tcpCh.Conn.LocalAddr()
}

func (tcpCh *TcpChannel) RemoteAddr() net.Addr {
	return tcpCh.Conn.RemoteAddr()
}

func (tcpCh *TcpChannel) NewPacket() gch.IPacket {
	w := &TcpPacket{}
	w.Packet = *gch.NewPacket(tcpCh, gch.NETWORK_TCP)
	return w
}

// TcpPacket Tcp包
type TcpPacket struct {
	gch.Packet
}
