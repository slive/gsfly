/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcpx

import (
	gch "github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
	gws "github.com/gorilla/websocket"
	"net"
	"time"
)

// WsChannel
type WsChannel struct {
	gch.Channel
	Conn   *gws.Conn
	params map[string]interface{}
}

func newWsChannel(parent interface{}, wsconn *gws.Conn, conf gch.IChannelConf, chHandle *gch.ChHandle, params map[string]interface{}, server bool) *WsChannel {
	ch := &WsChannel{Conn: wsconn, params: params}
	ch.Channel = *gch.NewDefChannel(parent, conf, chHandle, server)
	return ch
}

func NewWsSimpleChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, onReadHandler gch.ChHandleFunc, server bool) *WsChannel {
	chHandle := gch.NewDefChHandle(onReadHandler)
	return NewWsChannel(parent, wsConn, chConf, chHandle, nil, server)
}

// NewWsChannel 创建WsChannel
func NewWsChannel(parent interface{}, wsConn *gws.Conn, chConf gch.IChannelConf, chHandle *gch.ChHandle, params map[string]interface{}, server bool) *WsChannel {
	ch := newWsChannel(parent, wsConn, chConf, chHandle, params, server)
	wsConn.SetReadLimit(int64(chConf.GetReadBufSize()))
	ch.SetId(wsConn.LocalAddr().String() + "->" + wsConn.RemoteAddr().String())
	return ch
}

func (wsCh *WsChannel) Open() error {
	err := wsCh.StartChannel(wsCh)
	if err == nil {
		gch.HandleOnActive(gch.NewChHandleContext(wsCh, nil))
	}
	return err
}

func (wsCh *WsChannel) Close() {
	wsCh.StopChannel(wsCh)
}

func (wsCh *WsChannel) Read() (gch.IPacket, error) {
	// TODO 超时配置
	now := time.Now()
	conf := wsCh.GetConf()
	failTime := time.Duration(conf.GetCloseRevFailTime())
	duration := conf.GetReadTimeout() * time.Second * failTime
	// 一次失败都会失败
	wsCh.Conn.SetReadDeadline(now.Add(duration))
	msgType, data, err := wsCh.readMessage()
	if err != nil {
		logx.WarnTracef(wsCh, "read ws err:%v", err)
		gch.RevStatisFail(wsCh, now)
		return nil, err
	}

	wspacket := wsCh.NewPacket().(*WsPacket)
	wspacket.MsgType = msgType
	wspacket.SetData(data)
	gch.RevStatis(wspacket, true)
	return wspacket, err
}

func (wsCh *WsChannel) readMessage() (messageType int, p []byte, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			reerr, ok := rec.(error)
			if ok {
				err = reerr
			}
		}
	}()
	return wsCh.Conn.ReadMessage()
}

func (wsCh *WsChannel) IsReadLoopContinued(err error) bool {
	// 失败不继续
	return false
}

func (wsCh *WsChannel) Write(datapacket gch.IPacket) error {
	return wsCh.Channel.Write(datapacket)
	// if wsCh.IsClosed() {
	// 	return errors.New("wschannel had closed, chId:" + wsCh.GetId())
	// }
	//
	// chHandle := wsCh.GetChHandle()
	// defer func() {
	// 	rec := recover()
	// 	if rec != nil {
	// 		logx.Error("write ws error, chId:%v, error:%v", wsCh.GetId(), rec)
	// 		err, ok := rec.(error)
	// 		if !ok {
	// 			err = errors.New(fmt.Sprintf("%v", rec))
	// 		}
	// 		// 捕获处理消息异常
	// 		chHandle.OnErrorHandle(wsCh, common.NewError1(gch.ERR_WRITE, err))
	// 		// 有异常，终止执行
	// 		wsCh.StopChannel(wsCh)
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
	// 	conf := wsCh.GetConf()
	// 	// TODO 设置超时?
	// 	wsCh.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	// 	err := wsCh.Conn.WriteMessage(wspacket.MsgType, data)
	// 	if err != nil {
	// 		logx.Error("write ws error:", err)
	// 		gch.SendStatis(wspacket, false)
	// 		panic(err)
	// 		return err
	// 	}
	//
	// 	gch.SendStatis(wspacket, true)
	// 	logx.Info(wsCh.GetChStatis().StringSend())
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

func (wsCh *WsChannel) WriteByConn(datapacket gch.IPacket) error {
	wspacket := datapacket.(*WsPacket)
	data := wspacket.GetData()
	conf := wsCh.GetConf()
	// TODO 设置超时?
	wsCh.Conn.SetWriteDeadline(time.Now().Add(conf.GetWriteTimeout() * time.Second))
	err := wsCh.Conn.WriteMessage(wspacket.MsgType, data)
	if err != nil {
		logx.Error("write ws error:", err)
		gch.SendStatis(wspacket, false)
		panic(err)
		return err
	}
	return nil
}

// GetConn Deprecated
func (wsCh *WsChannel) GetConn() net.Conn {
	return wsCh.Conn.UnderlyingConn()
}

func (wsCh *WsChannel) LocalAddr() net.Addr {
	return wsCh.Conn.LocalAddr()
}

func (wsCh *WsChannel) RemoteAddr() net.Addr {
	return wsCh.Conn.RemoteAddr()
}

func (wsCh *WsChannel) GetParams() map[string]interface{} {
	return wsCh.params
}

// NewPacket 创建ws对应的packet默认TextMessage 文本类型
func (wsCh *WsChannel) NewPacket() gch.IPacket {
	w := &WsPacket{}
	w.Packet = *gch.NewPacket(wsCh, gch.NETWORK_WS)
	// 默认TextMessage 文本类型
	w.MsgType = gws.TextMessage
	return w
}

type WsPacket struct {
	gch.Packet

	// ws类型
	MsgType int
}
