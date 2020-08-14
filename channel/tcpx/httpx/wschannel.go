/*
 * Author:slive
 * DATE:2020/7/17
 */
package httpx

import (
	gws "github.com/gorilla/websocket"
	gch "gsfly/channel"
	logx "gsfly/logger"
	"net"
	"time"
)

type WsChannel struct {
	gch.Channel
	Conn *gws.Conn
}

func newWsChannel(parent interface{}, wsconn *gws.Conn, conf gch.IChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := &WsChannel{Conn: wsconn}
	ch.Channel = *gch.NewDefChannel(parent, conf, chHandle)
	return ch
}

func NewWsSimpleChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, msgFunc gch.OnMsgHandle) *WsChannel {
	chHandle := gch.NewDefChHandle(msgFunc)
	return NewWsChannel(parent, wsConn, chConf, chHandle)
}

func NewWsChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, chHandle *gch.ChannelHandle) *WsChannel {
	ch := newWsChannel(parent, wsConn, chConf, chHandle)
	wsConn.SetReadLimit(int64(chConf.GetReadBufSize()))
	ch.SetId("ws-" + wsConn.LocalAddr().String() + "-" + wsConn.RemoteAddr().String())
	return ch
}

func (b *WsChannel) Start() error {
	return b.StartChannel(b)
}

func (b *WsChannel) Stop() {
	b.StopChannel(b)
}

func (b *WsChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	// conf := b.GetChConf()
	now := time.Now()
	// b.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	msgType, data, err := b.Conn.ReadMessage()
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

func (b *WsChannel) Write(datapacket gch.IPacket) error {
	return b.Channel.Write(datapacket)
	// if b.IsClosed() {
	// 	return errors.New("wschannel had closed, chId:" + b.GetId())
	// }
	//
	// chHandle := b.GetChHandle()
	// defer func() {
	// 	rec := recover()
	// 	if rec != nil {
	// 		logx.Error("write ws error, chId:%v, error:%v", b.GetId(), rec)
	// 		err, ok := rec.(error)
	// 		if !ok {
	// 			err = errors.New(fmt.Sprintf("%v", rec))
	// 		}
	// 		// 捕获处理消息异常
	// 		chHandle.OnErrorHandle(b, common.NewError1(gch.ERR_WRITE, err))
	// 		// 有异常，终止执行
	// 		b.StopChannel(b)
	// 	}
	// }()
	//
	// if datapacket.IsPrepare() {
	// 	// 发送前的处理
	// 	befWriteHandle := chHandle.OnBefWriteHandle
	// 	if befWriteHandle != nil {
	// 		err := befWriteHandle(datapacket)
	// 		if err != nil{
	// 			logx.Error("befWriteHandle error:", err)
	// 			return err
	// 		}
	// 	}
	// 	wspacket := datapacket.(*WsPacket)
	// 	data := wspacket.GetData()
	// 	conf := b.GetChConf()
	// 	// TODO 设置超时?
	// 	b.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	// 	err := b.Conn.WriteMessage(wspacket.MsgType, data)
	// 	if err != nil {
	// 		logx.Error("write ws error:", err)
	// 		gch.SendStatis(wspacket, false)
	// 		panic(err)
	// 		return err
	// 	}
	//
	// 	gch.SendStatis(wspacket, true)
	// 	logx.Info(b.GetChStatis().StringSend())
	// 	// 发送成功后的处理
	// 	aftWriteHandle := chHandle.OnAftWriteHandle
	// 	if aftWriteHandle != nil {
	// 		aftWriteHandle(datapacket)
	// 	}
	// 	return err
	// } else {
	// 	logx.Warn("datapacket is not prepare.")
	// }
	// return nil
}

func (b *WsChannel) WriteByConn(datapacket gch.IPacket) error {
	wspacket := datapacket.(*WsPacket)
	data := wspacket.GetData()
	conf := b.GetChConf()
	// TODO 设置超时?
	b.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	err := b.Conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		logx.Error("write ws error:", err)
		gch.SendStatis(wspacket, false)
		panic(err)
		return err
	}
	return nil
}

// GetConn Deprecated
func (b *WsChannel) GetConn() net.Conn {
	return b.Conn.UnderlyingConn()
}

func (b *WsChannel) IsReadLoopContinued(err error) bool {
	return false
}

func (b *WsChannel) LocalAddr() net.Addr {
	return b.Conn.LocalAddr()
}

func (b *WsChannel) RemoteAddr() net.Addr {
	return b.Conn.RemoteAddr()
}

func (b *WsChannel) NewPacket() gch.IPacket {
	w := &WsPacket{}
	w.Packet = *gch.NewPacket(b, gch.PROTOCOL_WS)
	return w
}

type WsPacket struct {
	gch.Packet
	MsgType int
}
