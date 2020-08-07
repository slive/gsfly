/*
 * Author:slive
 * DATE:2020/7/17
 */
package ws

import (
	"errors"
	"fmt"
	gws "github.com/gorilla/websocket"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"gsfly/util"
	"net"
	"time"
)

type WsChannel struct {
	gch.BaseChannel
	conn *gws.Conn
}

func newWsChannel(wsconn *gws.Conn, conf gch.ChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := &WsChannel{conn: wsconn}
	ch.BaseChannel = *gch.NewDefaultBaseChannel(conf, chHandle)
	return ch
}

func NewWsSimpleChannel(wsConn *gws.Conn, chConf gch.ChannelConf, msgFunc gch.OnMsgHandle) *WsChannel {
	chHandle := gch.NewDefaultChHandle(msgFunc)
	return NewWsChannel(wsConn, chConf, chHandle)
}

func NewWsChannel(wsConn *gws.Conn, chConf gch.ChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := newWsChannel(wsConn, chConf, chHandle)
	wsConn.SetReadLimit(int64(chConf.GetReadBufSize()))
	ch.SetChId("ws-" + wsConn.LocalAddr().String() + "-" + wsConn.RemoteAddr().String())
	return ch
}

func (b *WsChannel) Read() (gch.Packet, error) {
	// TODO 超时配置
	// conf := b.GetChConf()
	now := time.Now()
	// b.conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	msgType, data, err := b.conn.ReadMessage()
	if err != nil {
		logx.Warn("read ws err:", err)
		gch.RevStatisFail(b, now)
		return nil, err
	}

	wspacket := b.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	gch.RevStatis(wspacket, true)
	logx.Info(b.GetChStatis().StringRev())
	return wspacket, err
}

func (b *WsChannel) Write(packet gch.Packet) error {
	if b.IsClosed() {
		return errors.New("wschannel had closed, chId:" + b.GetChId())
	}

	defer func() {
		rec := recover()
		if rec != nil {
			logx.Error("write ws error, chId:%v, error:%v", b.GetChId(), rec)
			err, ok := rec.(error)
			if !ok {
				err = errors.New(fmt.Sprintf("%v", rec))
			}
			// 捕获处理消息异常
			b.GetChHandle().OnErrorHandle(b, util.NewError1(gch.ERR_WRITE, err))
			// 有异常，终止执行
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
		logx.Error("write ws error:", err)
		gch.SendStatis(wspacket, false)
		panic(err)
		return err
	}

	gch.SendStatis(wspacket, true)
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
