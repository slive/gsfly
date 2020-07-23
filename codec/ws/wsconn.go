/*
 * Author:slive
 * DATE:2020/7/17
 */
package ws

import (
	gws "github.com/gorilla/websocket"
	gconn "gsfly/conn"
)

type WsConn struct {
	gconn.BaseConn
	conn *gws.Conn
}

func NewWsConn(wsconn *gws.Conn) *WsConn {
	ch := &WsConn{conn: wsconn}
	ch.BaseConn = *gconn.NewBaseConn()
	return ch
}

func StartWsConn(wsconn *gws.Conn, msgFunc gconn.HandleMsgFunc) *WsConn {
	ch := NewWsConn(wsconn)
	ch.SetHandleMsgFunc(msgFunc)
	go ch.ReadLoop()
	return ch
}

func (b *WsConn) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *WsConn) Read(packet gconn.Packet) error {
	// TODO 设置超时？
	msgType, data, err := b.conn.ReadMessage()
	if err != nil {
		return err
	}
	wspacket := packet.(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	return err
}

func (b *WsConn) ReadLoop() error {
	defer b.Close()
	for {
		rev := NewWsPacket()
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

func (b *WsConn) Write(packet gconn.Packet) error {
	// TODO 设置超时？
	wspacket := packet.(*WsPacket)
	data := wspacket.GetData()
	err := b.conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		return err
	}
	return err
}

func NewWsPacket() *WsPacket {
	w := &WsPacket{}
	w.Basepacket = *gconn.NewBasePacket(gconn.PROTOCOL_WS)
	return w
}

type WsPacket struct {
	gconn.Basepacket
	MsgType int
}
