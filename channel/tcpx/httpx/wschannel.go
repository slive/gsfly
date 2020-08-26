/*
 * Author:slive
 * DATE:2020/7/17
 */
package httpx

import (
	gch "github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
	gws "github.com/gorilla/websocket"
	"net"
	"time"
)

type WsChannel struct {
	gch.Channel
	Conn   *gws.Conn
	params map[string]interface{}
}

func newWsChannel(parent interface{}, wsconn *gws.Conn, conf gch.IChannelConf, chHandle *gch.ChannelHandle, params map[string]interface{}) *WsChannel {
	ch := &WsChannel{Conn: wsconn, params: params}
	ch.Channel = *gch.NewDefChannel(parent, conf, chHandle)
	return ch
}

func NewWsSimpleChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, msgFunc gch.OnMsgHandle) *WsChannel {
	chHandle := gch.NewDefChHandle(msgFunc)
	return NewWsChannel(parent, wsConn, chConf, chHandle, nil)
}

func NewWsChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, chHandle *gch.ChannelHandle, params map[string]interface{}) *WsChannel {
	ch := newWsChannel(parent, wsConn, chConf, chHandle, params)
	wsConn.SetReadLimit(int64(chConf.GetReadBufSize()))
	ch.SetId("ws-" + wsConn.LocalAddr().String() + "-" + wsConn.RemoteAddr().String())
	return ch
}

func (wsChannel *WsChannel) Start() error {
	err := wsChannel.StartChannel(wsChannel)
	if err == nil {
		onRegisterHandle := wsChannel.GetChHandle().OnRegisteredHandle
		if onRegisterHandle != nil {
			onRegisterHandle(wsChannel, nil)
		}
	}
	return err
}

func (wsChannel *WsChannel) Stop() {
	if !wsChannel.IsClosed() {
		onUnRegisterHandle := wsChannel.GetChHandle().OnUnRegisteredHandle
		if onUnRegisterHandle != nil {
			onUnRegisterHandle(wsChannel, nil)
		}
	}
	wsChannel.StopChannel(wsChannel)
}

func (wsChannel *WsChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	// conf := wsChannel.GetConf()
	now := time.Now()
	// wsChannel.Conn.SetReadDeadline(now.Add(conf.GetReadTimeout() * time.Second))
	msgType, data, err := wsChannel.Conn.ReadMessage()
	if err != nil {
		logx.Warn("read ws err:", err)
		gch.RevStatisFail(wsChannel, now)
		return nil, err
	}

	wspacket := wsChannel.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	gch.RevStatis(wspacket, true)
	return wspacket, err
}

func (wsChannel *WsChannel) Write(datapacket gch.IPacket) error {
	return wsChannel.Channel.Write(datapacket)
	// if wsChannel.IsClosed() {
	// 	return errors.New("wschannel had closed, chId:" + wsChannel.GetId())
	// }
	//
	// chHandle := wsChannel.GetChHandle()
	// defer func() {
	// 	rec := recover()
	// 	if rec != nil {
	// 		logx.Error("write ws error, chId:%v, error:%v", wsChannel.GetId(), rec)
	// 		err, ok := rec.(error)
	// 		if !ok {
	// 			err = errors.New(fmt.Sprintf("%v", rec))
	// 		}
	// 		// 捕获处理消息异常
	// 		chHandle.OnErrorHandle(wsChannel, common.NewError1(gch.ERR_WRITE, err))
	// 		// 有异常，终止执行
	// 		wsChannel.StopChannel(wsChannel)
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
	// 	conf := wsChannel.GetConf()
	// 	// TODO 设置超时?
	// 	wsChannel.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	// 	err := wsChannel.Conn.WriteMessage(wspacket.MsgType, data)
	// 	if err != nil {
	// 		logx.Error("write ws error:", err)
	// 		gch.SendStatis(wspacket, false)
	// 		panic(err)
	// 		return err
	// 	}
	//
	// 	gch.SendStatis(wspacket, true)
	// 	logx.Info(wsChannel.GetChStatis().StringSend())
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

func (wsChannel *WsChannel) WriteByConn(datapacket gch.IPacket) error {
	wspacket := datapacket.(*WsPacket)
	data := wspacket.GetData()
	conf := wsChannel.GetConf()
	// TODO 设置超时?
	wsChannel.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	err := wsChannel.Conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		logx.Error("write ws error:", err)
		gch.SendStatis(wspacket, false)
		panic(err)
		return err
	}
	return nil
}

// GetConn Deprecated
func (wsChannel *WsChannel) GetConn() net.Conn {
	return wsChannel.Conn.UnderlyingConn()
}

func (wsChannel *WsChannel) LocalAddr() net.Addr {
	return wsChannel.Conn.LocalAddr()
}

func (wsChannel *WsChannel) RemoteAddr() net.Addr {
	return wsChannel.Conn.RemoteAddr()
}

func (wsChannel *WsChannel) GetParams() map[string]interface{} {
	return wsChannel.params
}

func (wsChannel *WsChannel) NewPacket() gch.IPacket {
	w := &WsPacket{}
	w.Packet = *gch.NewPacket(wsChannel, gch.PROTOCOL_WS)
	return w
}

type WsPacket struct {
	gch.Packet
	MsgType int
}
