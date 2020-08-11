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

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 15,
	ReadBufferSize:   10 * 1024,
	WriteBufferSize:  10 * 1024,
}

type HttpWsServer struct {
	BaseServer
	ServerConf   *HttpxServerConf
	msgHandlers  map[string]*gch.ChannelHandle
	httpHandlers map[string]HttpHandleFunc
}

// Http和Websocket 的服务监听
func NewHttpxServer(serverConf *HttpxServerConf) Server {
	t := &HttpWsServer{
		httpHandlers: make(map[string]HttpHandleFunc),
		msgHandlers:  make(map[string]*gch.ChannelHandle),
	}
	t.BaseCommunication = *NewCommunication(nil)
	t.ServerConf = serverConf
	t.Channels = make(map[string]gch.Channel, 10)
	return t
}

// AddHttpHandleFunc 添加http处理方法
func (t *HttpWsServer) AddHttpHandleFunc(pattern string, httpHandleFunc HttpHandleFunc) {
	t.httpHandlers[pattern] = httpHandleFunc
}

// AddWsHandleFunc 添加Websocket处理方法
func (t *HttpWsServer) AddWsHandleFunc(pattern string, wsHandleFunc *gch.ChannelHandle) {
	t.msgHandlers[pattern] = wsHandleFunc
}

// GetChannelHandle 暂时不支持，用GetMsgHandlers()代替
func (t *HttpWsServer) GetChannelHandle() *gch.ChannelHandle {
	logx.Panic("unsupport")
	return nil
}

func (t *HttpWsServer) GetMsgHandlers() map[string]*gch.ChannelHandle {
	return t.msgHandlers
}

func (t *HttpWsServer) Start() error {
	if !t.Closed {
		return errors.New("server had opened, id:" + t.GetId())
	}

	// http处理事件
	httpHandlers := t.httpHandlers
	if httpHandlers != nil {
		for key, f := range httpHandlers {
			http.HandleFunc(key, f)
		}
	}

	defer func() {
		t.Stop()
	}()

	wsHandlers := t.GetMsgHandlers()
	if wsHandlers != nil {
		// ws处理事件
		acceptChannels := t.Channels
		for key, f := range wsHandlers {
			http.HandleFunc(key, func(writer http.ResponseWriter, r *http.Request) {
				logx.Info("requestWs:", r.URL)
				err := startWs(writer, r, upgrader, t.ServerConf, f, acceptChannels)
				if err != nil {
					logx.Error("start ws error:", err)
				}
			})
		}
	}

	addr := t.ServerConf.GetAddrStr()
	logx.Info("start http listen, addr:", addr)
	err := http.ListenAndServe(addr, nil)
	t.Closed = false
	if err != nil {
		for {
			select {
			case <-t.Exit:
				break
			}
		}
	}
	return err
}

type HttpHandleFunc func(http.ResponseWriter, *http.Request)

// startWs 启动ws处理
func startWs(w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader, serverConf *HttpxServerConf, handle *gch.ChannelHandle, acceptChannels map[string]gch.Channel) error {
	connLen := len(acceptChannels)
	maxAcceptSize := serverConf.GetMaxChannelSize()
	if connLen >= maxAcceptSize {
		return errors.New("max accept size:" + fmt.Sprintf("%v", maxAcceptSize))
	}

	// upgrade处理
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Println("upgrade error:", err)
		return err
	}
	if err != nil {
		logx.Error("accept error:", nil)
		return err
	}

	wsCh := httpx.NewWsChannel(conn, serverConf, handle)
	err = wsCh.Start()
	if err == nil {
		// TODO 线程安全？
		acceptChannels[wsCh.GetChId()] = wsCh
	}
	return err
}

type KcpServer struct {
	BaseServer
	ServerConf *KcpServerConf
}

func NewKcpServer(kcpServerConf *KcpServerConf, chHandle *gch.ChannelHandle) Server {
	k := &KcpServer{
		ServerConf: kcpServerConf,
	}
	k.BaseCommunication = *NewCommunication(chHandle)
	k.Channels = make(map[string]gch.Channel, 10)
	return k
}

func (k *KcpServer) Start() error {
	if !k.Closed {
		return errors.New("server had opened, id:" + k.GetId())
	}

	kcpServerConf := k.ServerConf
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kcp addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcp error, addr:", addr, err)
		return err
	}

	kwsChannels := k.Channels
	defer func() {
		for key, kch := range kwsChannels {
			kch.Stop()
			delete(kwsChannels, key)
		}
	}()

	for {
		kcpConn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpconn error:", nil)
			return err
		}

		kcpCh := kcpx.NewKcpChannel(kcpConn, kcpServerConf, k.ChannelHandle)
		err = kcpCh.Start()
		if err == nil {
			kwsChannels[kcpCh.GetChId()] = kcpCh
		}
	}
	return nil
}

type Kws00Server struct {
	KcpServer
	onKwsMsgHandle kcpx.OnKws00MsgHandle
}

func NewKws00Server(kcpServerConf *KcpServerConf, onKwsMsgHandle kcpx.OnKws00MsgHandle,
	onRegisterHandle gch.OnRegisterHandle, onUnRegisterHandle gch.OnUnRegisterHandle) Server {
	k := &Kws00Server{}
	k.ServerConf = kcpServerConf
	chHandle := kcpx.NewKws00Handle(onRegisterHandle, onUnRegisterHandle)
	k.BaseCommunication = *NewCommunication(chHandle)
	k.onKwsMsgHandle = onKwsMsgHandle
	k.Channels = make(map[string]gch.Channel, 10)
	return k
}

func (k *Kws00Server) Start() error {
	if !k.Closed {
		return errors.New("server had opened, id:" + k.GetId())
	}

	kcpServerConf := k.ServerConf
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kws00 addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kws00 error, addr:", addr, err)
		return err
	}

	kwsChannels := k.Channels
	defer func() {
		for key, kch := range kwsChannels {
			kch.Stop()
			delete(kwsChannels, key)
		}
	}()

	for {
		kcpConn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpconn error:", nil)
			return err
		}

		chHandle := k.ChannelHandle
		kcpCh := kcpx.NewKws00Channel(kcpConn, &kcpServerConf.BaseChannelConf, k.onKwsMsgHandle, chHandle)
		kcpCh.ChannelHandle = chHandle
		err = kcpCh.Start()
		if err == nil {
			kwsChannels[kcpCh.GetChId()] = kcpCh
		}
	}
	return nil
}

type UdpServer struct {
	BaseServer
	ServerConf *UdpServerConf
}

func NewUdpServer(serverConf *UdpServerConf, channelHandle *gch.ChannelHandle) Server {
	k := &UdpServer{
		ServerConf: serverConf,
	}
	k.BaseCommunication = *NewCommunication(channelHandle)
	k.Channels = make(map[string]gch.Channel, 10)
	return k
}

func (u *UdpServer) Start() error {
	serverConf := u.ServerConf
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
	ch := udpx.NewUdpChannel(conn, serverConf, u.ChannelHandle)
	err = ch.Start()
	return err
}
