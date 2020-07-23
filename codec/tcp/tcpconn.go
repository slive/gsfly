/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcp

import (
	logx "github.com/sirupsen/logrus"
	gconn "gsfly/conn"
	"net"
)

type TcpConn struct {
	gconn.BaseConn
	conn *net.TCPConn
}

func NewTcpConn(conn *net.TCPConn) *TcpConn {
	ch := &TcpConn{conn: conn}
	ch.BaseConn = *gconn.NewBaseConn()
	return ch
}

func StartTcpConn(conn *net.TCPConn, msgFunc gconn.HandleMsgFunc) *TcpConn {
	ch := NewTcpConn(conn)
	ch.SetHandleMsgFunc(msgFunc)
	go ch.ReadLoop()
	return ch
}

func (b *TcpConn) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

// TODO 配置化？
var readbf = make([]byte, 10*1024)

func (b *TcpConn) Read(datapack gconn.Packet) error {
	// TODO 超时？
	//b.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		return err
	}

	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	logx.Info("receive:", string(bytes))
	return err
}

func (b *TcpConn) ReadLoop() error {
	defer b.Close()
	for {
		rev := NewTcpPacket()
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

func (b *TcpConn) Write(datapack gconn.Packet) error {
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

func NewTcpPacket() *TcpPacket {
	w := &TcpPacket{}
	w.Basepacket = *gconn.NewBasePacket(gconn.PROTOCOL_TCP)
	return w
}

type TcpPacket struct {
	gconn.Basepacket
}
