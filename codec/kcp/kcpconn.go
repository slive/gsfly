/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcp

import (
	"fmt"
	logx "github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
	gconn "gsfly/conn"
)

type KcpConn struct {
	gconn.BaseConn
	conn *kcp.UDPSession
}

func NewKcpConn(kcpconn *kcp.UDPSession) *KcpConn {
	ch := &KcpConn{conn: kcpconn}
	ch.BaseConn = *gconn.NewBaseConn()
	return ch
}

func StartKcpConn(kcpconn *kcp.UDPSession, msgFunc gconn.HandleMsgFunc) *KcpConn {
	ch := NewKcpConn(kcpconn)
	ch.SetHandleMsgFunc(handlerMessage)
	// TODO 配置化
	go ch.ReadLoop()
	return ch
}

func (b *KcpConn) GetChId() string {
	return b.conn.RemoteAddr().String() + ":" + fmt.Sprintf("%s", b.conn.GetConv())
}

func (b *KcpConn) IsOpen() bool {
	return true
}

// TODO 配置化？
var readbf = make([]byte, 10*1024)

func (b *KcpConn) Read(datapack gconn.Packet) error {
	// TODO 超时？
	//b.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	readNum, err := b.conn.Read(readbf)
	if err != nil {
		return err
	}
	// 接收到8个字节数据，是bug?
	if readNum <= 8 {
		return nil
	}
	bytes := readbf[0:readNum]
	datapack.SetData(bytes)
	logx.Info("receive:", string(bytes))
	return err
}

func (b *KcpConn) ReadLoop() error {
	defer b.Close()
	for {
		rev := NewKcpPacket()
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

func (b *KcpConn) Write(datapack gconn.Packet) error {
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

func handlerMessage(conn gconn.Conn, datapack gconn.Packet) error {
	logx.Println("handler datapack:", datapack)
	return nil
}

type KcpPacket struct {
	gconn.Basepacket
}

func NewKcpPacket() *KcpPacket {
	k := &KcpPacket{}
	k.SetPType(gconn.PROTOCOL_KCP)
	return k
}
