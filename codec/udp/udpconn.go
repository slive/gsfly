/*
 * Author:slive
 * DATE:2020/7/17
 */
package udp

import (
	logx "github.com/sirupsen/logrus"
	gconn "gsfly/conn"
	"net"
)

type UdpConn struct {
	gconn.BaseConn
	conn *net.UDPConn
}

func NewUdpConn(conn *net.UDPConn) *UdpConn {
	ch := &UdpConn{conn: conn}
	ch.BaseConn = *gconn.NewBaseConn()
	return ch
}

func StartUdpConn(conn *net.UDPConn, msgFunc gconn.HandleMsgFunc) *UdpConn {
	ch := NewUdpConn(conn)
	ch.SetHandleMsgFunc(msgFunc)
	go ch.ReadLoop()
	return ch
}

func (b *UdpConn) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

// TODO 配置化？
var readbf = make([]byte, 10*1024)

func (b *UdpConn) Read(datapack gconn.Packet) error {
	// TODO 超时？
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		return err
	}

	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	logx.Info("receive:", string(bytes))
	return err
}

func (b *UdpConn) ReadLoop() error {
	defer b.Close()
	for {
		rev := NewUdpPacket()
		err := b.Read(rev)
		if err != nil {
			return err
		}
		if rev.GetPrepare() {
			msgFunc := b.GetHandleMsgFunc()
			msgFunc(b, rev)
		}
	}
}

func (b *UdpConn) Write(datapack gconn.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover error:", i)
			b.Close()
		}
	}()
	if datapack.GetPrepare() {
		bytes := datapack.GetData()
		_, err := b.conn.Write(bytes)
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

func NewUdpPacket() *UdpPacket {
	w := &UdpPacket{}
	w.Basepacket = *gconn.NewBasePacket(gconn.PROTOCOL_UDP)
	return w
}

type UdpPacket struct {
	gconn.Basepacket
}
