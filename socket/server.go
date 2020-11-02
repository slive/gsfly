/*
 * 基于TCP协议的，如TCP，Http和Websocket 的服务监听
 * Author:slive
 * DATE:2020/7/17
 */
package socket

import (
	"errors"
	"fmt"
	gch "github.com/Slive/gsfly/channel"
	"github.com/Slive/gsfly/channel/tcpx"
	kcpx "github.com/Slive/gsfly/channel/udpx/kcpx"
	logx "github.com/Slive/gsfly/logger"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	http "net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

// IServerListener 服务监听接口
type IServerListener interface {
	ISocket
	GetConf() IServerConf
	GetChannelPool() *hashmap.Map
	Listen() error

	// GetHttpServer 针对http
	GetHttpServer() *http.Server
	SetHttpServer(httpServer *http.Server)
}

// ServerListener 服务监听
type ServerListener struct {
	Socket
	Conf        IServerConf
	ChannelPool *hashmap.Map
	httpServer  *http.Server
}

// NewServerListener 创建服务监听器
// parent 父类
// serverConf 服务端配置
// chHandle handle
func NewServerListener(parent interface{}, serverConf IServerConf, chHandle gch.IChHandle) *ServerListener {
	if chHandle == nil {
		errMsg := "chHandle is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	if serverConf == nil {
		errMsg := "serverConf is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	b := &ServerListener{
		Conf:        serverConf,
		ChannelPool: hashmap.New(),
	}
	b.Socket = *NewSocketConn(parent, chHandle, nil)
	return b
}

func (listener *ServerListener) GetId() string {
	return listener.Conf.GetAddrStr()
}

func (listener *ServerListener) Close() {
	if !listener.Closed {
		id := listener.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop listen, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop listen, id:", id)
		listener.Closed = true
		listener.Exit <- true
		acceptChannels := listener.GetChannelPool().Values()
		for _, ch := range acceptChannels {
			ch.(gch.IChannel).Stop()
		}
	}
}

func (listener *ServerListener) GetChannelPool() *hashmap.Map {
	return listener.ChannelPool
}

func (listener *ServerListener) GetConf() IServerConf {
	return listener.Conf
}

// Listen 监听方法，可自定义实现
func (listener *ServerListener) Listen() error {
	network := listener.GetConf().GetNetwork()
	switch network {
	case gch.NETWORK_WS:
		return listenWs(listener)
	case gch.NETWORK_KCP:
		return listenKcp(listener)
	default:
		logx.Info("unsupport network:", network)
		return nil
	}
	return errors.New("start serverListener error.")
}

const KEY_HTTP_REQUEST = "http-request"

// Http和Websocket 的服务监听
// parent 父类，可选
// serverConf 服务器配置，必须项
// chHandle channel处理类，必须项
// httpServer http监听服务，可选，为空时，根据serverConf的ip/port进行创建监听
func listenWs(serverListener *ServerListener) error {
	id := serverListener.GetId()
	if !serverListener.IsClosed() {
		return errors.New("httpServer had opened, id:" + id)
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.Warnf("finish listenWs, id:%v, ret:%v", id, ret)
			serverListener.Close()
		} else {
			logx.Info("finish listenWs, id:", id)
		}
	}()

	wsServerConf := serverListener.GetConf().(IWsServerConf)
	// ws依赖http做升级而来
	upgrader := websocket.Upgrader{
		HandshakeTimeout: wsServerConf.GetReadTimeout() * time.Second,
		ReadBufferSize:   wsServerConf.GetReadBufSize(),
		WriteBufferSize:  wsServerConf.GetWriteBufSize(),
	}

	httpServer := serverListener.GetHttpServer()
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
			logx.Info("listenAnServe id:", id)
			err := httpServer.ListenAndServe()
			if err != nil {
				logx.Error("listenAnServe error:", err)
				panic(err)
			}
		}()
		go func() {
			// 关闭操作
			s := make(chan os.Signal, 1)
			signal.Notify(s)
			select {
			case sg := <-s:
				httpServer.Close()
				logx.Infof("listenAnServe close, id:%v, signal:%v", id, sg)
			}
		}()

		// ws处理事件
		http.HandleFunc(wsServerConf.GetPath(), func(writer http.ResponseWriter, req *http.Request) {
			logx.Info("requestWs:", req.URL)
			err := upgradeWs(serverListener, writer, req, upgrader)
			if err != nil {
				logx.Error("start ws error:", err)
			}
		})
	} else {
		httpHandler := httpServer.Handler
		httpServer.Handler = newProxyHandler(httpHandler, upgrader, serverListener)
	}
	serverListener.Closed = false
	return nil
}

func newProxyHandler(handler http.Handler, upgrader websocket.Upgrader, serverListener IServerListener) *proxyHandler {
	return &proxyHandler{handler: handler, upgrader: upgrader, serverListener: serverListener}
}

type proxyHandler struct {
	handler        http.Handler
	upgrader       websocket.Upgrader
	serverListener IServerListener
}

func (proxy *proxyHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	logx.Info("request:", uri)
	conf := proxy.serverListener.GetConf().(IWsServerConf)
	if strings.Contains(uri, conf.GetPath()) {
		err := upgradeWs(proxy.serverListener, writer, req, proxy.upgrader)
		if err != nil {
			logx.Error("start ws error:", err)
		}
		return
	}

	proxy.handler.ServeHTTP(writer, req)
}

// listenWs 启动ws处理
func upgradeWs(serverListener IServerListener, writer http.ResponseWriter, req *http.Request, upgr websocket.Upgrader) error {
	acceptChannels := serverListener.GetChannelPool()
	serverConf := serverListener.GetConf().(IWsServerConf)
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
	logx.Infof("form:%v, params:%v, url:%v", form, params, urlStr)
	// upgrade处理
	subPros := serverConf.GetSubProtocol()
	header := req.Header
	if subPros != nil && len(subPros) > 0 {
		// 设置subprotocol
		for _, pro := range subPros {
			header.Add("Sec-WebSocket-Protocol", pro)
			logx.Info("set ws subprotocol:", pro)
		}
	}
	conn, err := upgr.Upgrade(writer, req, header)
	if err != nil {
		logx.Println("upgrade error:", err)
		return err
	}
	addHttpRequest(serverListener, req)

	chHandle := serverListener.GetChHandle().(*gch.ChHandle)
	// OnStopHandle重新包装，以便释放资源
	chHandle.SetOnInActive(ConverOnInActiveHandler(serverListener.GetChannelPool(), chHandle.GetOnInActive()))
	// 复制新的handle
	// chHandle = gch.CopyChannelHandle(chHandle)
	wsCh := tcpx.NewWsChannel(serverListener, conn, serverConf, chHandle, params, true)
	err = wsCh.Start()
	if err == nil {
		// TODO 线程安全？
		acceptChannels.Put(wsCh.GetId(), wsCh)
	}
	return err
}

func addHttpRequest(serverListener IServerListener, req *http.Request) {
	serverListener.AddAttach(KEY_HTTP_REQUEST, req)
}

func (serverListener *ServerListener) GetHttpServer() *http.Server {
	return serverListener.httpServer
}

func (serverListener *ServerListener) SetHttpServer(httpServer *http.Server) {
	serverListener.httpServer = httpServer
}

// ConverOnInActiveHandle 转化OnStopHandle方法
func ConverOnInActiveHandler(channels *hashmap.Map, onInActiveHandler gch.ChHandleFunc) func(ctx gch.IChHandleContext) {
	return func(ctx gch.IChHandleContext) {
		// 释放现有资源
		chId := ctx.GetChannel().GetId()
		channels.Remove(chId)
		logx.Infof("remove serverchannel, chId:%v, channelSize:%v", chId, channels.Size())
		if onInActiveHandler != nil {
			onInActiveHandler(ctx)
		}
	}
}

func listenKcp(serverListener *ServerListener) error {
	if !serverListener.IsClosed() {
		return errors.New("server had opened, id:" + serverListener.GetId())
	}

	kcpServerConf := serverListener.GetConf()
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
			logx.Warnf("finish kcp serverListener, id:%v, ret:%v", serverListener.GetId(), ret)
			serverListener.Close()
		} else {
			logx.Info("finish kcp serverListener, id:", serverListener.GetId())
		}
	}()

	kcpChannels := serverListener.GetChannelPool()
	go func() {
		for {
			kcpConn, err := list.AcceptKCP()
			if err != nil {
				logx.Error("accept kcpconn error:", nil)
				panic(err)
			}
			chHandle := serverListener.GetChHandle().(*gch.ChHandle)
			// OnStopHandle重新包装，以便释放资源
			chHandle.SetOnInActive(ConverOnInActiveHandler(kcpChannels, chHandle.GetOnInActive()))
			// 复制新的handle
			// chHandle = gch.CopyChannelHandle(chHandle)
			kcpCh := kcpx.NewKcpChannel(serverListener, kcpConn, kcpServerConf, chHandle, true)
			err = kcpCh.Start()
			if err == nil {
				kcpChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()

	if err == nil {
		serverListener.Closed = false
	}

	return nil
}