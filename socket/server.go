/*
 * 基于TCP协议的，如TCP，Http和Websocket 的服务监听
 * Author:slive
 * DATE:2020/7/17
 */
package socket

import (
	"errors"
	"fmt"
	gch "github.com/slive/gsfly/channel"
	"github.com/slive/gsfly/channel/tcpx"
	"github.com/slive/gsfly/channel/udpx"
	kcpx "github.com/slive/gsfly/channel/udpx/kcpx"
	logx "github.com/slive/gsfly/logger"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"net"
	http "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// IServerSocket 服务监听接口
type IServerSocket interface {
	ISocket
	GetConf() IServerConf
	GetChannels() *hashmap.Map
	Listen() error

	// GetHttpServer 针对http
	GetHttpServer() *http.Server
	SetHttpServer(httpServer *http.Server)

	// GetBasePath 监听基本path，符合该规则（有优先级控制）匹配的requestPath都可以进来
	GetBasePath() string
}

// ServerSocket 服务监听
type ServerSocket struct {
	Socket
	Conf       IServerConf
	channels   *hashmap.Map
	httpServer *http.Server
	basePath   string
}

// NewServerSocket 创建服务监听器
// parent 父类
// serverConf 服务端配置
// chHandle handle
func NewServerSocket(parent interface{}, serverConf IServerConf, chHandle gch.IChHandle) *ServerSocket {
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

	b := &ServerSocket{
		Conf:     serverConf,
		channels: hashmap.New(),
	}
	b.Socket = *NewSocket(parent, chHandle, nil)
	b.SetId("server#" + b.Conf.GetNetwork().String() + "#" + b.Conf.GetAddrStr())
	return b
}

func (serverSocket *ServerSocket) Close() {
	if !serverSocket.Closed {
		defer func() {
			ret := recover()
			logx.InfoTracef(serverSocket, "finish to stop listen, ret:%v", ret)
		}()
		logx.InfoTracef(serverSocket, "start to stop listen.")
		serverSocket.Closed = true
		serverSocket.Exit <- true
		acceptChannels := serverSocket.GetChannels().Values()
		for _, ch := range acceptChannels {
			ch.(gch.IChannel).Release()
		}
	}
}

func (serverSocket *ServerSocket) GetChannels() *hashmap.Map {
	return serverSocket.channels
}

func (serverSocket *ServerSocket) GetConf() IServerConf {
	return serverSocket.Conf
}
func (serverSocket *ServerSocket) GetHttpServer() *http.Server {
	return serverSocket.httpServer
}

func (serverSocket *ServerSocket) SetHttpServer(httpServer *http.Server) {
	serverSocket.httpServer = httpServer
}

func (serverSocket *ServerSocket) GetBasePath() string {
	return serverSocket.basePath
}

// ConverOnInActiveHandle 转化OnStopHandle方法
func ConverOnInActiveHandler(channels *hashmap.Map, onInActiveHandler gch.ChHandleFunc) func(ctx gch.IChHandleContext) {
	return func(ctx gch.IChHandleContext) {
		// 释放现有资源
		chId := ctx.GetChannel().GetId()
		channels.Remove(chId)
		logx.InfoTracef(ctx, "remove serverchannel, channelSize:%v", channels.Size())
		if onInActiveHandler != nil {
			onInActiveHandler(ctx)
		}
	}
}

// Listen 监听方法，可自定义实现
func (serverSocket *ServerSocket) Listen() error {
	network := serverSocket.GetConf().GetNetwork()
	switch network {
	case gch.NETWORK_WS:
		return listenWs(serverSocket)
	case gch.NETWORK_HTTP:
		// TODO http也是ws处理
		return listenWs(serverSocket)
	case gch.NETWORK_KCP:
		return listenKcp(serverSocket)
	case gch.NETWORK_TCP:
		return listenTcp(serverSocket)
	case gch.NETWORK_UDP:
		return listenUdp(serverSocket)
	default:
		logx.InfoTracef(serverSocket, "unsupport network:%v", network)
		return nil
	}
	return errors.New("start serverSocket error, id:" + serverSocket.GetId())
}

const KEY_HTTP_REQUEST = "http-request"

// Http和Websocket 的服务监听
// parent 父类，可选
// serverConf 服务器配置，必须项
// chHandle channel处理类，必须项
// httpServer http监听服务，可选，为空时，根据serverConf的ip/port进行创建监听
func listenWs(ss *ServerSocket) error {
	id := ss.GetId()
	if !ss.IsClosed() {
		return errors.New("websocket Server is opened, id:" + id)
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.WarnTracef(ss, "finish listenWs, ret:%v", ret)
			ss.Close()
		} else {
			logx.InfoTracef(ss, "finish listenWs.")
		}
	}()

	wsServerConf := ss.GetConf().(IWsServerConf)
	// ws依赖http做升级而来
	upgrader := websocket.Upgrader{
		HandshakeTimeout: wsServerConf.GetReadTimeout() * time.Second,
		ReadBufferSize:   wsServerConf.GetReadBufSize(),
		WriteBufferSize:  wsServerConf.GetWriteBufSize(),
	}

	addrStr := wsServerConf.GetAddrStr()
	httpServer := ss.GetHttpServer()
	if httpServer == nil {
		// 为空时，根据serverConf的ip/port进行创建监听
		httpServer = &http.Server{
			Addr:              addrStr,
			TLSConfig:         nil,
			ReadTimeout:       wsServerConf.GetReadTimeout() * time.Second,
			ReadHeaderTimeout: wsServerConf.GetReadTimeout() * time.Second,
			WriteTimeout:      wsServerConf.GetWriteTimeout() * time.Second,
			IdleTimeout:       wsServerConf.GetReadTimeout() * time.Second * 3,
			MaxHeaderBytes:    1 << 20,
		}
		// 启动监听
		go func() {
			// 异步监听http和ws
			logx.InfoTracef(ss, "listenAnServe addrStr:%v", addrStr)
			err := httpServer.ListenAndServe()
			if err != nil {
				logx.ErrorTracef(ss, "listenAnServe error:%v", err)
				panic(err)
			}
		}()

		// 等待关闭操作
		go func() {
			s := make(chan os.Signal, 1)
			signal.Notify(s, os.Kill, os.Interrupt, syscall.SIGABRT, syscall.SIGTERM)
			select {
			case sg := <-s:
				httpServer.Close()
				logx.InfoTracef(ss, "listenAnServe close, signal:%v", sg)
			}
		}()

		// 处理children
		wsChildren := wsServerConf.GetListenConfs()
		if wsChildren != nil {
			for _, child := range wsChildren {
				network := child.GetNetwork()
				if len(network) <= 0 {
					// 子配置没有配置network，则取父节点
					network = wsServerConf.GetNetwork()
				}
				if network == gch.NETWORK_WS {
					// ws处理事件，针对不同的basePath进行处理
					http.HandleFunc(child.GetBasePath(), func(writer http.ResponseWriter, req *http.Request) {
						logx.InfoTracef(ss, "requestWs:%v", req.URL)
						err := upgradeWs(ss, writer, req, upgrader, child)
						if err != nil {
							logx.ErrorTracef(ss, "start ws error:%v", err)
						}
					})
				}
			}
		}
	} else {
		// 已在外面完成的监听，重写httphandler处理事件，以便通过不同的path分别处理http和ws
		httpHandler := httpServer.Handler
		httpServer.Handler = newProxyHandler(httpHandler, upgrader, ss)
	}
	ss.Closed = false
	return nil
}

func newProxyHandler(handler http.Handler, upgrader websocket.Upgrader, serverListener IServerSocket) *proxyHandler {
	return &proxyHandler{handler: handler, upgrader: upgrader, serverSocket: serverListener}
}

type proxyHandler struct {
	handler      http.Handler
	upgrader     websocket.Upgrader
	serverSocket IServerSocket
}

// ServeHTTP 通过不同的path分别处理http和ws
func (proxy *proxyHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	serverSocket := proxy.serverSocket
	logx.InfoTracef(serverSocket, "http request:%v", uri)
	conf := serverSocket.GetConf().(IWsServerConf)
	// TODO 模糊匹配?，优先级如何？
	wsChildren := conf.GetListenConfs()
	if wsChildren != nil {
		for _, child := range wsChildren {
			path := child.GetBasePath()
			// 优先处理ws的handler
			if strings.Contains(uri, path) {
				err := upgradeWs(serverSocket, writer, req, proxy.upgrader, child)
				if err != nil {
					logx.ErrorTracef(serverSocket, "start ws error:%v", err)
				}
				return
			}
		}
	}

	// 除了ws，默认处理http的handler
	proxy.handler.ServeHTTP(writer, req)
}

// listenWs 启动ws处理
func upgradeWs(ss IServerSocket, writer http.ResponseWriter, req *http.Request, upgr websocket.Upgrader, childConf IServerChildConf) error {
	acceptChannels := ss.GetChannels()
	serverConf := ss.GetConf().(IWsServerConf)
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
	logx.InfoTracef(ss, "form:%v, params:%v, url:%v", form, params, urlStr)
	// upgrade处理
	subPros := childConf.GetAttach(WS_SUBPROTOCOL_KEY)
	header := req.Header
	if subPros != nil {
		// 设置subprotocol
		pros, ok := subPros.(string)
		if ok {
			header.Add("Sec-WebSocket-Protocol", pros)
			logx.Info("set ws subprotocol:", pros)
		} else {
			pros, ok := subPros.([]string)
			if ok {
				for _, pro := range pros {
					header.Add("Sec-WebSocket-Protocol", pro)
					logx.Info("set ws subprotocol:", pro)
				}
			}
		}

	}
	conn, err := upgr.Upgrade(writer, req, header)
	if err != nil {
		logx.ErrorTracef(ss, "upgrade error:%v", err)
		return err
	}
	addHttpRequest(ss, req)

	chHandle := ss.GetChHandle().(*gch.ChHandle)
	// OnInActiveHandle重新包装，以便释放资源
	chHandle.SetOnRelease(ConverOnInActiveHandler(acceptChannels, chHandle.GetOnRelease()))
	wsCh := tcpx.NewWsChannel(ss, conn, serverConf, chHandle, params, true)
	// 设置为请求过来的path
	wsCh.SetRelativePath(req.URL.Path)
	err = wsCh.Open()
	if err == nil {
		// TODO 线程安全？
		acceptChannels.Put(wsCh.GetId(), wsCh)
	}
	return err
}

func addHttpRequest(serverListener IServerSocket, req *http.Request) {
	serverListener.AddAttach(KEY_HTTP_REQUEST, req)
}

func listenKcp(ss *ServerSocket) error {
	if !ss.IsClosed() {
		return errors.New("kcp server is opened, id:" + ss.GetId())
	}

	kcpServerConf := ss.GetConf()
	addr := kcpServerConf.GetAddrStr()
	logx.InfoTracef(ss, "listen kcp addr:%v", addr)
	listKcp, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.ErrorTracef(ss, "listen kcp error:%v", err)
		return err
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.ErrorTracef(ss, "finish kcp serverSocket, error:%v", ret)
			ss.Close()
		} else {
			logx.InfoTracef(ss, "finish kcp serverSocket.")
		}
	}()

	kcpChannels := ss.GetChannels()
	go func() {
		for {
			kcpConn, err := listKcp.AcceptKCP()
			if err != nil {
				logx.ErrorTracef(ss, "accept kcp error:%v", err)
				listKcp.Close()
				panic(err)
			}

			schHandle := ss.GetChHandle().(*gch.ChHandle)
			// 复制一份handle，避免相互覆盖
			chHandle := gch.CopyChHandle(schHandle)
			// OnInActiveHandle重新包装，以便释放资源
			chHandle.SetOnRelease(ConverOnInActiveHandler(kcpChannels, chHandle.GetOnRelease()))
			kcpCh := kcpx.NewKcpChannel(ss, kcpConn, kcpServerConf, chHandle, true)
			err = kcpCh.Open()
			if err == nil {
				kcpChannels.Put(kcpCh.GetId(), kcpCh)
			}
		}
	}()

	if err == nil {
		ss.Closed = false
	}
	return err
}

func listenTcp(ss *ServerSocket) error {
	if !ss.IsClosed() {
		return errors.New("server had opened, id:" + ss.GetId())
	}

	serverConf := ss.GetConf()
	addr := serverConf.GetAddrStr()
	logx.InfoTracef(ss, "listen tcp addr:%v", addr)
	network := serverConf.GetNetwork().String()
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	listenTCP, err := net.ListenTCP(network, tcpAddr)
	if err != nil {
		logx.InfoTracef(ss, "listen tcp error:%v", err)
		return err
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.ErrorTracef(ss, "finish tcp serverSocket, err:%v", ret)
			ss.Close()
		} else {
			logx.InfoTracef(ss, "finish tcp serverSocket.")
		}
	}()

	channels := ss.GetChannels()
	go func() {
		for {
			tcpConn, err := listenTCP.AcceptTCP()
			if err != nil {
				logx.ErrorTracef(ss, "accept tcp error:%v", err)
				listenTCP.Close()
				panic(err)
			}
			chHandle := ss.GetChHandle().(*gch.ChHandle)
			// OnInActiveHandle重新包装，以便释放资源
			chHandle.SetOnRelease(ConverOnInActiveHandler(channels, chHandle.GetOnRelease()))
			tcpCh := tcpx.NewTcpChannel(ss, tcpConn, serverConf, chHandle, true)
			err = tcpCh.Open()
			if err == nil {
				channels.Put(tcpCh.GetId(), tcpCh)
			}
		}
	}()

	return err
}

func listenUdp(ss *ServerSocket) error {
	if !ss.IsClosed() {
		return errors.New("udp server is opened, id:" + ss.GetId())
	}

	serverConf := ss.GetConf()
	addr := serverConf.GetAddrStr()
	logx.InfoTracef(ss, "listen udp addr:%v", addr)
	network := serverConf.GetNetwork().String()
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	udpConn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		logx.ErrorTracef(ss, "listen udp error:%v", err)
		return err
	}

	defer func() {
		ret := recover()
		if ret != nil {
			logx.ErrorTracef(ss, "finish udp serverSocket, err:%v", ret)
			ss.Close()
		} else {
			logx.InfoTracef(ss, "finish udp serverSocket.")
		}
	}()

	readBufSize := serverConf.GetReadBufSize()
	if readBufSize > udpx.Max_UDP_Buf {
		readBufSize = udpx.Max_UDP_Buf
	}
	readbf := make([]byte, readBufSize)
	channels := ss.GetChannels()
	go func() {
		for {
			// TODO 是否有性能问题？
			readNum, addr, err := udpConn.ReadFromUDP(readbf)
			if readNum < 0 {
				continue
			}
			if err != nil {
				logx.ErrorTracef(ss, "read udp error:%v", err)
				udpConn.Close()
				panic(err)
			}

			var udpCh *udpx.UdpChannel
			buf := readbf[0:readNum]
			udpChId := udpx.FetchUdpId(udpConn, addr)
			channel, found := channels.Get(udpChId)
			if !found {
				// 第一次生成一个channel
				chHandle := ss.GetChHandle().(*gch.ChHandle)
				// OnInActiveHandle重新包装，以便释放资源
				chHandle.SetOnRelease(ConverOnInActiveHandler(channels, chHandle.GetOnRelease()))
				udpCh = udpx.NewUdpChannel(ss, udpConn, serverConf, chHandle, addr, true)
				err = udpCh.Open()
				if err == nil {
					channels.Put(udpCh.GetId(), udpCh)
				}
				udpCh.CacheServerRead(buf)
			} else {
				// 已存在直接缓存
				udpCh = channel.(*udpx.UdpChannel)
			}

			if udpCh != nil {
				// 通过缓存
				udpCh.CacheServerRead(buf)
			}
		}
	}()

	if err == nil {
		ss.Closed = false
	}
	return err
}
