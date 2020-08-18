/*
 * 基于TCP协议的，如TCP，Http和Websocket 的服务监听
 * Author:slive
 * DATE:2020/7/17
 */
package bootstrap

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	httpx "gsfly/channel/tcpx/httpx"
	udpx "gsfly/channel/udpx"
	kcpx "gsfly/channel/udpx/kcpx"
	logx "gsfly/logger"
	"net"
	http "net/http"
	"time"
)

type IWsServerStrap interface {
	IServerStrap

	GetHttpServer() *http.Server

	GetHttpRequest() *http.Request
}

type WsServerStrap struct {
	ServerStrap
	httpServer  *http.Server
	httpRequest *http.Request
}

// Http和Websocket 的服务监听
// parent 父类，可选
// serverConf 服务器配置，必须项
// chHandle channel处理类，必须项
// httpServer http监听服务，可选，为空时，根据serverConf的ip/port进行创建监听
func NewWsServerStrap(parent interface{}, serverConf IWsServerConf, chHandle *gch.ChannelHandle, httpServer *http.Server) *WsServerStrap {
	t := &WsServerStrap{}
	t.ServerStrap = *NewServerStrap(parent, serverConf, chHandle)
	t.httpServer = httpServer
	return t
}

func (wsServerStrap *WsServerStrap) Start() error {
	if !wsServerStrap.Closed {
		return errors.New("httpServer had opened, id:" + wsServerStrap.GetId())
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.Warnf("finish httpws serverstrap, id:%v, ret:%v", wsServerStrap.GetId(), ret)
			wsServerStrap.Stop()
		} else {
			logx.Info("finish httpws serverstrap, id:", wsServerStrap.GetId())
		}
	}()

	wsServerConf := wsServerStrap.GetConf().(IWsServerConf)
	httpServer := wsServerStrap.httpServer
	if httpServer == nil {
		// 为空时，根据serverConf的ip/port进行创建监听
		httpServer = &http.Server{
			Addr:              wsServerConf.GetAddrStr(),
			TLSConfig:         nil,
			ReadTimeout:       wsServerConf.GetReadTimeout() * time.Second,
			ReadHeaderTimeout: wsServerConf.GetReadTimeout() * time.Second,
			WriteTimeout:      wsServerConf.GetWriteTimeout() * time.Second,
			IdleTimeout:       wsServerConf.GetReadTimeout() * time.Second * 3,
			MaxHeaderBytes:    1 << 20,
		}
		go func() {
			err := httpServer.ListenAndServe()
			if err != nil {
				logx.Error("listenAnServe error:", err)
				wsServerStrap.Stop()
			}
		}()
	}

	upgrader := websocket.Upgrader{
		HandshakeTimeout: wsServerConf.GetReadTimeout() * time.Second,
		ReadBufferSize:   wsServerConf.GetReadBufSize(),
		WriteBufferSize:  wsServerConf.GetWriteBufSize(),
	}

	// ws处理事件
	http.HandleFunc(wsServerConf.GetPath(), func(writer http.ResponseWriter, req *http.Request) {
		logx.Info("requestWs:", req.URL)
		err := wsServerStrap.startWs(writer, req, upgrader)
		if err != nil {
			logx.Error("start ws error:", err)
		}
	})
	wsServerStrap.Closed = false
	return nil
}

// startWs 启动ws处理
func (wsServerStrap *WsServerStrap) startWs(writer http.ResponseWriter, req *http.Request, upgr websocket.Upgrader) error {
	acceptChannels := wsServerStrap.GetChannels()
	serverConf := wsServerStrap.GetConf().(IWsServerConf)
	connLen := acceptChannels.Size()
	maxAcceptSize := serverConf.GetMaxChannelSize()
	if connLen >= maxAcceptSize {
		return errors.New("max accept size:" + fmt.Sprintf("%v", maxAcceptSize))
	}

	params := make(map[string]interface{})
	req.ParseForm()
	form := req.Form
	for key, val := range form {
		params[key] = val[0]
	}
	urlStr := req.URL.String()
	logx.Infof("params:%v, url:%v", params, urlStr)
	// upgrade处理
	conn, err := upgr.Upgrade(writer, req, nil)
	if err != nil {
		logx.Println("upgrade error:", err)
		return err
	}
	wsServerStrap.httpRequest = req

	wsCh := httpx.NewWsChannel(wsServerStrap, conn, serverConf, wsServerStrap.GetChHandle(), params)
	err = wsCh.Start()
	if err == nil {
		// TODO 线程安全？
		acceptChannels.Put(wsCh.GetId(), wsCh)
	}
	return err
}

func (wsServerStrap *WsServerStrap) GetHttpServer() *http.Server {
	return wsServerStrap.httpServer
}

type KcpServerStrap struct {
	ServerStrap
}

func NewKcpServer(parent interface{}, kcpServerConf IKcpServerConf, chHandle *gch.ChannelHandle) IServerStrap {
	k := &KcpServerStrap{}
	k.ServerStrap = *NewServerStrap(parent, kcpServerConf, chHandle)
	k.Conf = kcpServerConf
	return k
}

func (k *KcpServerStrap) Start() error {
	if !k.Closed {
		return errors.New("server had opened, id:" + k.GetId())
	}

	kcpServerConf := k.GetConf()
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kcp addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcp error, addr:", addr, err)
		return err
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.Warnf("finish kcp serverstrap, id:%v, ret:%v", k.GetId(), ret)
			k.Stop()
		} else {
			logx.Info("finish kcp serverstrap, id:", k.GetId())
		}
	}()

	kwsChannels := k.Channels
	go func() {
		for {
			kcpConn, err := list.AcceptKCP()
			if err != nil {
				logx.Error("accept kcpconn error:", nil)
				panic(err)
			}
			// OnStopHandle重新包装，以便释放资源
			handle := k.ChannelHandle
			handle.OnStopHandle = ConverOnStopHandle(k.Channels, handle.OnStopHandle)
			kcpCh := kcpx.NewKcpChannel(k, kcpConn, kcpServerConf, handle)
			err = kcpCh.Start()
			if err == nil {
				kwsChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()

	if err == nil {
		k.Closed = false
	}

	return nil
}

type Kws00ServerStrap struct {
	KcpServerStrap
}

func NewKws00Server(parent interface{}, kcpServerConf IKw00ServerConf, onKwsMsgHandle gch.OnMsgHandle,
	onRegisterHandle gch.OnRegisterHandle, onUnRegisterHandle gch.OnUnRegisterHandle) IServerStrap {
	k := &Kws00ServerStrap{}
	k.Conf = kcpServerConf
	chHandle := kcpx.NewKws00Handle(onKwsMsgHandle, onRegisterHandle, onUnRegisterHandle)
	k.ServerStrap = *NewServerStrap(parent, kcpServerConf, chHandle)
	return k
}

func (k *Kws00ServerStrap) Start() error {
	if !k.Closed {
		return errors.New("server had opened, id:" + k.GetId())
	}

	kcpServerConf := k.GetConf()
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kws00 addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kws00 error, addr:", addr, err)
		return err
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.Warnf("finish kws00 serverstrap, id:%v, ret:%v", k.GetId(), ret)
			k.Stop()
		} else {
			logx.Info("finish kws00 serverstrap, id:", k.GetId())
		}
	}()

	kwsChannels := k.Channels
	go func() {
		for {
			kcpConn, err := list.AcceptKCP()
			if err != nil {
				logx.Error("accept kcpconn error:", nil)
				panic(err)
			}

			chHandle := k.ChannelHandle
			kcpCh := kcpx.NewKws00Channel(k, kcpConn, kcpServerConf, chHandle, nil)
			err = kcpCh.Start()
			if err == nil {
				kwsChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()
	if err == nil {
		k.Closed = false
	}
	return nil
}

type UdpServerStrap struct {
	ServerStrap
}

func NewUdpServer(parent interface{}, serverConf IUdpServerConf, channelHandle *gch.ChannelHandle) IServerStrap {
	k := &UdpServerStrap{}
	k.ServerStrap = *NewServerStrap(parent, serverConf, channelHandle)
	k.Conf = serverConf
	return k
}

func (u *UdpServerStrap) Start() error {
	serverConf := u.GetConf()
	addr := serverConf.GetAddrStr()
	logx.Info("dial udp addr:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.Error("resolve updaddr error:", err)
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logx.Error("listen upd error:", err)
		return err
	}

	// TODO udp有源和目标地址之分，待实现
	ch := udpx.NewUdpChannel(u, conn, serverConf, u.ChannelHandle)
	err = ch.Start()
	if err == nil {
		u.Closed = false
	}
	return err
}
