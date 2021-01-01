/*
 * 客户端 connection
 * Author:slive
 * DATE:2020/9/21
 */
package socket

import (
	"errors"
	"fmt"
	gch "github.com/Slive/gsfly/channel"
	"github.com/Slive/gsfly/channel/tcpx"
	"github.com/Slive/gsfly/channel/udpx"
	"github.com/Slive/gsfly/channel/udpx/kcpx"
	logx "github.com/Slive/gsfly/logger"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"net"
)

// IClientSocket 客户端conn
type IClientSocket interface {
	ISocket
	GetConf() IClientConf
	GetChannel() gch.IChannel
	Dial() error
}

type ClientSocket struct {
	Socket

	// 路径
	reqPath string

	Conf IClientConf
	// 每一个客户端只有一个channel
	Channel gch.IChannel
}

// NewClientSocket 创建客户端socketconn
// parent 父类
// clientConf 客户端配置
// chHandle handle
// inputParams 所需参数
func NewClientSocket(parent interface{}, clientConf IClientConf, handle gch.IChHandle, inputParams map[string]interface{}) *ClientSocket {
	b := &ClientSocket{}
	b.Conf = clientConf
	b.Socket = *NewSocket(parent, handle, inputParams)
	b.SetId("client#" + b.Conf.GetNetwork().String() + "#" + b.Conf.GetAddrStr())
	return b
}

func (clientSocket *ClientSocket) GetChannel() gch.IChannel {
	return clientSocket.Channel
}

func (clientSocket *ClientSocket) GetConf() IClientConf {
	return clientSocket.Conf
}

func (clientSocket *ClientSocket) Dial() error {
	network := clientSocket.GetConf().GetNetwork()
	switch network {
	case gch.NETWORK_WS:
		return dialWs(clientSocket)
	case gch.NETWORK_KCP:
		return dialKcp(clientSocket)
	case gch.NETWORK_TCP:
		return dialTcp(clientSocket)
	case gch.NETWORK_HTTP:
		// TODO 默认为ws
		return dialWs(clientSocket)
	case gch.NETWORK_UDP:
		return dialUdp(clientSocket)
	default:
		return nil
	}
	return errors.New("start clentstrap error.")
}

// dialWs 拨号实现ws
func dialWs(cs *ClientSocket) error {
	wsClientConf := cs.GetConf().(IWsClientConf)
	handle := cs.GetChHandle().(*gch.ChHandle)

	url := wsClientConf.GetUrl()
	params := cs.GetInputParams()
	if params != nil && len(params) > 0 {
		path := wsClientConf.GetReqPath()
		if len(path) <= 0 {
			// 如果配置中的path为空，则选用参数中的path
			path := params["path"].(string)
			url = wsClientConf.GetUrlByPath(path)
		}
		url += "?"
		index := 1
		pLen := len(params)
		for key, v := range params {
			if index < pLen {
				url += key + "=" + fmt.Sprintf("%v&", v)
			} else {
				url += key + "=" + fmt.Sprintf("%v", v)
			}
			index++
		}
	}
	logx.InfoTracef(cs, "dial ws url:%v", url)
	conn, response, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logx.Error("dial ws error:", err)
		return err
	}

	// TODO 处理resonse？
	logx.InfoTracef(cs, "ws response:%v", response)
	// 创建channel
	wsCh := tcpx.NewWsChannel(cs, conn, wsClientConf, handle, params, false)
	wsCh.SetRelativePath(wsClientConf.GetReqPath())
	err = wsCh.Open()
	if err == nil {
		cs.Channel = wsCh
		cs.Closed = false
	}
	return err
}

func dialTcp(cs *ClientSocket) error {
	tcpClientConf := cs.GetConf()
	chHandle := cs.GetChHandle().(*gch.ChHandle)
	addr := tcpClientConf.GetAddrStr()
	logx.InfoTracef(cs, "dial tcp addr:%v", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logx.ErrorTracef(cs, "dial tcp addr error:%v", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logx.ErrorTracef(cs, "dial tcp error:%v", err)
		return err
	}

	tcpCh := tcpx.NewTcpChannel(cs, conn, tcpClientConf, chHandle, false)
	var path string
	params := cs.GetInputParams()
	if params != nil {
		p := params["path"]
		if p != nil {
			path = p.(string)
		}
	}
	tcpCh.SetRelativePath(path)
	err = tcpCh.Open()
	if err == nil {
		cs.Channel = tcpCh
		cs.Closed = false
	}
	return err
}

func dialKcp(cs *ClientSocket) error {
	kcpClientConf := cs.GetConf()
	chHandle := cs.GetChHandle().(*gch.ChHandle)
	addr := kcpClientConf.GetAddrStr()
	logx.InfoTracef(cs, "dial kcp addr:%v", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.ErrorTracef(cs, "dial kcp conn error:%v", err)
		return err
	}
	kcpCh := kcpx.NewKcpChannel(cs, conn, kcpClientConf, chHandle, false)
	var path string
	params := cs.GetInputParams()
	if params != nil {
		p := params["path"]
		if p != nil {
			path = p.(string)
		}
	}
	kcpCh.SetRelativePath(path)
	err = kcpCh.Open()
	if err == nil {
		cs.Channel = kcpCh
		cs.Closed = false
	}
	return err
}

func dialUdp(cs *ClientSocket) error {
	udpConf := cs.GetConf()
	chHandle := cs.GetChHandle().(*gch.ChHandle)
	addr := udpConf.GetAddrStr()
	logx.InfoTracef(cs, "dial udp addr:%v", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.ErrorTracef(cs, "dial udp addr error:%v", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logx.ErrorTracef(cs, "dial udp error:%v", err)
		return err
	}

	udpCh := udpx.NewUdpChannel(cs, conn, udpConf, chHandle, udpAddr, false)
	var path string
	params := cs.GetInputParams()
	if params != nil {
		p := params["path"]
		if p != nil {
			path = p.(string)
		}
	}
	udpCh.SetRelativePath(path)
	err = udpCh.Open()
	if err == nil {
		cs.Channel = udpCh
		cs.Closed = false
	}
	return err
}

func (clientSocket *ClientSocket) Close() {
	if !clientSocket.Closed {
		defer func() {
			ret := recover()
			logx.InfoTracef(clientSocket, "finish to stop client, ret:%v", ret)
		}()
		logx.InfoTracef(clientSocket, "start to stop client.")
		clientSocket.Closed = true
		clientSocket.Exit <- true
		clientSocket.Channel.Close()
		clientSocket.Channel = nil
		clientSocket.Channel.GetConn().Close()
	}
}
