/*
 * Author:slive
 * DATE:2020/7/17
 */
package ws

import (
	gws "github.com/gorilla/websocket"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type WsChannel struct {
	gch.BaseChannel
	conn *gws.Conn
}

func newWsChannel(wsconn *gws.Conn, conf *gch.BaseChannelConf) *WsChannel {
	ch := &WsChannel{conn: wsconn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	return ch
}

func NewWsChannel(wsConn *gws.Conn, chConf *gch.BaseChannelConf, msgFunc gch.HandleMsgFunc) *WsChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewWsChannelWithHandle(wsConn, chConf, chHandle)
}

func NewWsChannelWithHandle(wsConn *gws.Conn, chConf *gch.BaseChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := newWsChannel(wsConn, chConf)
	ch.ChannelHandle = *chHandle
	wsConn.SetReadLimit(int64(chConf.ReadBufSize))
	ch.SetChId("ws-" + wsConn.LocalAddr().String() + "-" + wsConn.RemoteAddr().String())
	return ch
}

func (b *WsChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	// conf := b.GetChConf()
	// b.conn.SetReadDeadline(time.Now().Add(conf.GetReadTimeout() * time.Second))
	msgType, data, err := b.conn.ReadMessage()
	if err != nil {
		logx.Info("read ws err:", err)
		return nil, err
	}

	wspacket := b.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	gch.RevStatis(wspacket)
	logx.Info(b.GetChStatis().StringRev())
	return wspacket, err
}

func (b *WsChannel) Write(packet gch.Packet) error {
	defer func() {
		i := recover()
		if i != nil {
			logx.Error("recover, write error:", i)
			b.StopChannel(b)
		}
	}()

	// TODO 设置超时？
	wspacket := packet.(*WsPacket)
	data := wspacket.GetData()
	conf := b.GetChConf()
	b.conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	err := b.conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		panic(err)
		return err
	}

	gch.SendStatis(wspacket)
	logx.Info(b.GetChStatis().StringSend())
	return err
}

func (b *WsChannel) StopChannel(channel gch.Channel) {
	if !b.IsClosed() {
		b.conn.Close()
	}
	b.BaseChannel.StopChannel(channel)
}

func (b *WsChannel) IsReadLoopContinued(err error) bool {
	return false
}

func (b *WsChannel) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}

func (b *WsChannel) RemoteAddr() net.Addr {
	return b.conn.RemoteAddr()
}

func (b *WsChannel) NewPacket() gch.Packet {
	w := &WsPacket{}
	w.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_WS)
	return w
}

type WsPacket struct {
	gch.Basepacket
	MsgType int
}
