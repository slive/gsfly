/*
 * 基于TCP协议的，如TCP，Http和Websocket 的服务监听
 * Author:slive
 * DATE:2020/7/17
 */
package bootstrap

import (
	"errors"
	"fmt"
	gch "github.com/Slive/gsfly/channel"
	"github.com/Slive/gsfly/channel/tcpx"
	httpx "github.com/Slive/gsfly/channel/tcpx/httpx"
	udpx "github.com/Slive/gsfly/channel/udpx"
	kcpx "github.com/Slive/gsfly/channel/udpx/kcpx"
	logx "github.com/Slive/gsfly/logger"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
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
			// 异步监听http和ws
			err := httpServer.ListenAndServe()
			if err != nil {
				logx.Error("listenAnServe error:", err)
				wsServerStrap.Stop()
			}
		}()
	}

	// ws依赖http做升级而来
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
	acceptChannels := wsServerStrap.GetChannelPool()
	serverConf := wsServerStrap.GetConf().(IWsServerConf)
	connLen := acceptChannels.Size()
	maxAcceptSize := serverConf.GetMaxChannelSize()
	if maxAcceptSize > 0 && connLen >= maxAcceptSize {
		return errors.New("max accept size:" + fmt.Sprintf("%v", maxAcceptSize))
	}

	// 拼接ws所需的parameter
	params := make(map[string]interface{})
	req.ParseForm()
	form := req.Form
	for key, val := range form {
		params[key] = val[0]
	}

	urlStr := req.URL.String()
	logx.Infof("params:%v, url:%v", params, urlStr)
	// upgrade处理
	subPros := serverConf.GetSubProtocol()
	header := req.Header
	if subPros != nil && len(subPros) > 0 {
		// 设置subprotocol
		for _, pro := range subPros {
			header.Add("Sec-WebSocket-Protocol", pro)
			logx.Info("set ws protocol:", pro)
		}
	}
	conn, err := upgr.Upgrade(writer, req, header)
	if err != nil {
		logx.Println("upgrade error:", err)
		return err
	}
	wsServerStrap.httpRequest = req

	chHandle := wsServerStrap.ChannelHandle
	// OnStopHandle重新包装，以便释放资源
	chHandle.OnStopHandle = ConverOnStopHandle(wsServerStrap.ChannelPool, chHandle.OnStopHandle)
	// 复制新的handle
	chHandle = gch.CopyChannelHandle(chHandle)
	wsCh := httpx.NewWsChannel(wsServerStrap, conn, serverConf, chHandle, params, true)
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

type ITcpServerStrap interface {
	IServerStrap
}
type TcpServerStrap struct {
	ServerStrap
}

func NewTcpServerStrap(parent interface{}, serverConf IUdpServerConf, channelHandle *gch.ChannelHandle) ITcpServerStrap {
	k := &TcpServerStrap{}
	k.ServerStrap = *NewServerStrap(parent, serverConf, channelHandle)
	k.Conf = serverConf
	return k
}

func (tcpSs *TcpServerStrap) Start() error {
	serverConf := tcpSs.GetConf()
	addr := serverConf.GetAddrStr()
	logx.Info("dial tcp addr:", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logx.Error("resolve tcpaddr error:", err)
		return err
	}

	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logx.Error("listen tcp error:", err)
		return err
	}
	go func() {
		conn, err := listen.AcceptTCP()
		if err != nil {
			logx.Error("listen tcp conn error:", err)
			tcpSs.Stop()
			return
		}

		chHandle := tcpSs.ChannelHandle
		// OnStopHandle重新包装，以便释放资源
		chHandle.OnStopHandle = ConverOnStopHandle(tcpSs.ChannelPool, chHandle.OnStopHandle)
		// 复制新的handle
		chHandle = gch.CopyChannelHandle(chHandle)
		ch := tcpx.NewTcpChannel(tcpSs, conn, serverConf, chHandle, true)
		err = ch.Start()
		if err == nil {
			tcpSs.GetChannelPool().Put(ch.GetId(), ch)
		}
	}()
	return err
}

type IKcpServerStrap interface {
	IServerStrap
}

type KcpServerStrap struct {
	ServerStrap
}

func NewKcpServerStrap(parent interface{}, kcpServerConf IKcpServerConf, chHandle *gch.ChannelHandle) IKcpServerStrap {
	k := &KcpServerStrap{}
	k.ServerStrap = *NewServerStrap(parent, kcpServerConf, chHandle)
	k.Conf = kcpServerConf
	return k
}

func (kcpServerStrap *KcpServerStrap) Start() error {
	if !kcpServerStrap.Closed {
		return errors.New("server had opened, id:" + kcpServerStrap.GetId())
	}

	kcpServerConf := kcpServerStrap.GetConf()
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
			logx.Warnf("finish kcp serverstrap, id:%v, ret:%v", kcpServerStrap.GetId(), ret)
			kcpServerStrap.Stop()
		} else {
			logx.Info("finish kcp serverstrap, id:", kcpServerStrap.GetId())
		}
	}()

	kwsChannels := kcpServerStrap.ChannelPool
	go func() {
		for {
			kcpConn, err := list.AcceptKCP()
			if err != nil {
				logx.Error("accept kcpconn error:", nil)
				panic(err)
			}
			chHandle := kcpServerStrap.GetChHandle()
			// OnStopHandle重新包装，以便释放资源
			chHandle.OnStopHandle = ConverOnStopHandle(kcpServerStrap.ChannelPool, chHandle.OnStopHandle)
			// 复制新的handle
			chHandle = gch.CopyChannelHandle(chHandle)
			kcpCh := kcpx.NewKcpChannel(kcpServerStrap, kcpConn, kcpServerConf, chHandle, true)
			err = kcpCh.Start()
			if err == nil {
				kwsChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()

	if err == nil {
		kcpServerStrap.Closed = false
	}

	return nil
}

type IKws00ServerStrap interface {
	IKcpServerStrap
}

type Kws00ServerStrap struct {
	KcpServerStrap
}

func NewKws00ServerStrap(parent interface{}, kcpServerConf IKw00ServerConf, onKwsMsgHandle gch.OnMsgHandle,
	onRegisterHandle gch.OnRegisteredHandle, onUnRegisterHandle gch.OnUnRegisteredHandle) IKws00ServerStrap {
	k := &Kws00ServerStrap{}
	k.Conf = kcpServerConf
	chHandle := kcpx.NewKws00Handle(onKwsMsgHandle, onRegisterHandle, onUnRegisterHandle)
	k.ServerStrap = *NewServerStrap(parent, kcpServerConf, chHandle)
	return k
}

func NewKws00ServerStrap1(parent interface{}, kcpServerConf IKw00ServerConf, chHandle *gch.ChannelHandle) IServerStrap {
	k := &Kws00ServerStrap{}
	k.Conf = kcpServerConf
	if chHandle.OnRegisteredHandle == nil {
		errMsg := "onRegisterhandle is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}
	k.ServerStrap = *NewServerStrap(parent, kcpServerConf, chHandle)
	return k
}

func (kws00ServefStrap *Kws00ServerStrap) Start() error {
	if !kws00ServefStrap.Closed {
		return errors.New("server had opened, id:" + kws00ServefStrap.GetId())
	}

	kcpServerConf := kws00ServefStrap.GetConf()
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
			logx.Warnf("finish kws00 serverstrap, id:%v, ret:%v", kws00ServefStrap.GetId(), ret)
			kws00ServefStrap.Stop()
		} else {
			logx.Info("finish kws00 serverstrap, id:", kws00ServefStrap.GetId())
		}
	}()

	kwsChannels := kws00ServefStrap.ChannelPool
	go func() {
		for {
			kcpConn, err := list.AcceptKCP()
			if err != nil {
				logx.Error("accept kcpconn error:", nil)
				panic(err)
			}

			logx.Info("accept kcpconn:", kcpConn.GetConv())

			chHandle := kws00ServefStrap.ChannelHandle
			newChHandle := gch.CopyChannelHandle(chHandle)
			// 复制新的handle
			// OnStopHandle重新包装，以便释放资源
			newChHandle.OnStopHandle = ConverOnStopHandle(kws00ServefStrap.ChannelPool, chHandle.OnStopHandle)
			kcpCh := kcpx.NewKws00Channel(kws00ServefStrap, kcpConn, kcpServerConf, newChHandle, nil, true)
			err = kcpCh.Start()
			if err == nil {
				kwsChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()
	if err == nil {
		kws00ServefStrap.Closed = false
	}
	return nil
}

type IUdpServerStrap interface {
	IServerStrap
}
type UdpServerStrap struct {
	ServerStrap
}

func NewUdpServerStrap(parent interface{}, serverConf IUdpServerConf, channelHandle *gch.ChannelHandle) IUdpServerStrap {
	k := &UdpServerStrap{}
	k.ServerStrap = *NewServerStrap(parent, serverConf, channelHandle)
	k.Conf = serverConf
	return k
}

func (udpServerStrap *UdpServerStrap) Start() error {
	serverConf := udpServerStrap.GetConf()
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
	chHandle := udpServerStrap.ChannelHandle
	// OnStopHandle重新包装，以便释放资源
	chHandle.OnStopHandle = ConverOnStopHandle(udpServerStrap.ChannelPool, chHandle.OnStopHandle)
	// 复制新的handle
	chHandle = gch.CopyChannelHandle(chHandle)
	ch := udpx.NewUdpChannel(udpServerStrap, conn, serverConf, chHandle, true)
	err = ch.Start()
	if err == nil {
		udpServerStrap.Closed = false
	}
	return err
}
