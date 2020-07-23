/*
 * Author:slive
 * DATE:2020/7/17
 */
package kcpws

import (
	logx "github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
	kcpx "gsfly/codec/kcp"
	"gsfly/codec/kcpws/frame"
	gconn "gsfly/conn"
)

type KwsConn struct {
	kcpx.KcpConn
	conn           *kcp.UDPSession
	handleKwsFrame HandleKwsFrame
}

func NewKwsConn(kcpconn *kcp.UDPSession) *KwsConn {
	ch := &KwsConn{conn: kcpconn}
	ch.KcpConn = *kcpx.NewKcpConn(kcpconn)
	return ch
}

func StartKwsConn(kcpconn *kcp.UDPSession, handleKwsFrame HandleKwsFrame) *KwsConn {
	ch := NewKwsConn(kcpconn)
	ch.handleKwsFrame = handleKwsFrame
	ch.SetHandleMsgFunc(handlerMessage)
	// TODO 配置化
	go ch.ReadLoop()
	return ch
}

func (b *KwsConn) ReadLoop() error {
	defer b.Close()
	for {
		rev := NewKwsPacket()
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

type KwsPacket struct {
	kcpx.KcpPacket
}

func NewKwsPacket() *KwsPacket {
	k := &KwsPacket{}
	k.SetPType(gconn.PROTOCOL_KWS)
	return k
}

func handlerMessage(conn gconn.Conn, datapack gconn.Packet) error {
	bf := frame.NewInputFrame(datapack.GetData())
	logx.Println("baseFrame:", bf.ToJsonString())
	kwsConn := conn.(*KwsConn)
	kwsConn.handleKwsFrame(conn, bf)
	return nil
}

type HandleKwsFrame func(conn gconn.Conn, frame frame.Frame) error
