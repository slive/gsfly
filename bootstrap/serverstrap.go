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
	"gsfly/channel/tcp/ws"
	"gsfly/channel/udp"
	kcpx "gsfly/channel/udp/kcp"
	logx "gsfly/logger"
	"net"
	httpx "net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 15,
	ReadBufferSize:   10 * 1024,
	WriteBufferSize:  10 * 1024,
}

type HttpWsServer struct {
	BaseServer
	msgHandlers  map[string]*gch.ChannelHandle
	httpHandlers map[string]HttpHandleFunc
}

// Http和Websocket 的服务监听
func NewServer(serverConf *BaseServerConf) Server {
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

func (tcpls *HttpWsServer) Start() error {
	if !tcpls.Closed {
		return errors.New("server had opened, id:" + tcpls.GetId())
	}

	// http处理事件
	httpHandlers := tcpls.httpHandlers
	if httpHandlers != nil {
		for key, f := range httpHandlers {
			httpx.HandleFunc(key, f)
		}
	}

	defer func() {
		tcpls.Stop()
	}()

	wsHandlers := tcpls.GetMsgHandlers()
	if wsHandlers != nil {
		// ws处理事件
		acceptChannels := tcpls.Channels
		for key, f := range wsHandlers {
			httpx.HandleFunc(key, func(writer httpx.ResponseWriter, r *httpx.Request) {
				logx.Info("requestWs:", r.URL)
				err := startWs(writer, r, tcpls.ServerConf, f, acceptChannels)
				if err != nil {
					logx.Error("start ws error:", err)
				}
			})
		}
	}

	addr := tcpls.ServerConf.GetAddrStr()
	logx.Info("start httpx listen, addr:", addr)
	err := httpx.ListenAndServe(addr, nil)
	tcpls.Closed = false
	if err != nil {
		for {
			select {
			case <-tcpls.Exit:
				break
			}
		}
	}
	return err
}

type HttpHandleFunc func(httpx.ResponseWriter, *httpx.Request)

// startWs 启动ws处理
func startWs(w httpx.ResponseWriter, r *httpx.Request, serverConf *BaseServerConf, handle *gch.ChannelHandle, acceptChannels map[string]gch.Channel) error {
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

	wsCh := ws.NewWsChannelWithHandle(conn, gch.Global_Conf.ChannelConf, handle)
	err = wsCh.StartChannel(wsCh)
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
			kch.StopChannel(kch)
			delete(kwsChannels, key)
		}
	}()

	for {
		kcpConn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpconn error:", nil)
			return err
		}

		kcpCh := kcpx.NewKcpChannelWithHandle(kcpConn, &kcpServerConf.BaseChannelConf, k.ChannelHandle)
		err = kcpCh.StartChannel(kcpCh)
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
	ch := udp.NewUdpChannelWithHandle(conn, &serverConf.BaseChannelConf, u.ChannelHandle)
	err = ch.StartChannel(ch)
	return err
}
