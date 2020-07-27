/*
 * Author:slive
 * DATE:2020/7/17
 */
package ws

import (
	gws "github.com/gorilla/websocket"
	gchannel "gsfly/channel"
	"gsfly/config"
	logx "gsfly/logger"
	"time"
)

type WsChannel struct {
	gchannel.BaseChannel
	conn *gws.Conn
}

func NewWsChannel(wsconn *gws.Conn, conf *config.ChannelConf) *WsChannel {
	ch := &WsChannel{conn: wsconn}
	ch.BaseChannel = *gchannel.NewBaseChannel(conf)
	return ch
}

func StartWsChannel(wsconn *gws.Conn, conf *config.ChannelConf, msgFunc gchannel.HandleMsgFunc) (*WsChannel, error) {
	return StartWsChannelWithHandle(wsconn, conf, gchannel.ChannelHandle{
		HandleMsgFunc: msgFunc,
	})
}

func StartWsChannelWithHandle(wsconn *gws.Conn, conf *config.ChannelConf, channelHandle gchannel.ChannelHandle) (*WsChannel, error) {
	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start wschannel error:", re)
		}
	}()

	var err error
	ch := NewWsChannel(wsconn, conf)
	ch.ChannelHandle = channelHandle
	wsconn.SetReadLimit(int64(conf.ReadBufSize))
	go gchannel.StartReadLoop(ch)
	handle := ch.ChannelHandle
	startFunc := handle.HandleStartFunc
	if startFunc != nil {
		err = startFunc(ch)
	}
	return ch, err
}

func (b *WsChannel) GetChId() string {
	return b.conn.LocalAddr().String() + ":" + b.conn.RemoteAddr().String()
}

func (b *WsChannel) Read() (packet gchannel.Packet, err error) {
	// TODO 超时配置
	conf := b.GetConf()
	b.conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	msgType, data, err := b.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	wspacket := b.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	return wspacket, err
}

func (b *WsChannel) Write(packet gchannel.Packet) error {
	// TODO 设置超时？
	wspacket := packet.(*WsPacket)
	data := wspacket.GetData()
	conf := b.GetConf()
	b.conn.SetWriteDeadline(time.Now().Add(conf.ReadTimeout * time.Second))
	err := b.conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		return err
	}
	return err
}

func (b *WsChannel) NewPacket() gchannel.Packet {
	w := &WsPacket{}
	w.Basepacket = *gchannel.NewBasePacket(b, gchannel.PROTOCOL_WS)
	return w
}

type WsPacket struct {
	gchannel.Basepacket
	MsgType int
}
