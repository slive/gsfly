/*
 * Author:slive
 * DATE:2020/7/17
 */
package ws

import (
	gws "github.com/gorilla/websocket"
	gch "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"time"
)

type WsChannel struct {
	gch.BaseChannel
	conn *gws.Conn
}

func newWsChannel(wsconn *gws.Conn, conf *config.ChannelConf) *WsChannel {
	ch := &WsChannel{conn: wsconn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf)
	return ch
}

func NewWsChannel(wsConn *gws.Conn, chConf *config.ChannelConf, msgFunc gch.HandleMsgFunc) *WsChannel {
	chHandle := gch.NewChHandle(msgFunc, nil, nil)
	return NewWsChannelWithHandle(wsConn, chConf, chHandle)
}

func NewWsChannelWithHandle(wsConn *gws.Conn, chConf *config.ChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := newWsChannel(wsConn, chConf)
	ch.ChannelHandle = *chHandle
	wsConn.SetReadLimit(int64(chConf.ReadBufSize))
	return ch
}

func (b *WsChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *WsChannel) Read() (packet gch.Packet, err error) {
	// TODO 超时配置
	conf := b.GetChConf()
	b.conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	msgType, data, err := b.conn.ReadMessage()
	if err != nil {
		logx.Info("receive err:", err)
		return nil, err
	}

	wspacket := b.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	logx.Info("receive ws:", string(data))
	gch.RevStatis(wspacket)

	return wspacket, err
}

func (b *WsChannel) Write(packet gch.Packet) error {
	// TODO 设置超时？
	wspacket := packet.(*WsPacket)
	data := wspacket.GetData()
	conf := b.GetChConf()
	b.conn.SetWriteDeadline(time.Now().Add(conf.WriteTimeout * time.Second))
	err := b.conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		return err
	}
	logx.Info("write ws:", string(data))
	gch.SendStatis(wspacket)
	return err
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
