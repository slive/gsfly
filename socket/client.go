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
	"github.com/Slive/gsfly/channel/udpx/kcpx"
	logx "github.com/Slive/gsfly/logger"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"net"
)

// IClientConn 客户端conn
type IClientConn interface {
	ISocket
	GetConf() IClientConf
	GetChannel() gch.IChannel
	Dial() error
}

type ClientConn struct {
	Socket

	// 路径
	reqPath string

	Conf IClientConf
	// 每一个客户端只有一个channel
	Channel gch.IChannel
}

// NewClientConn 创建客户端socketconn
// parent 父类
// clientConf 客户端配置
// chHandle handle
// inputParams 所需参数
func NewClientConn(parent interface{}, clientConf IClientConf, handle gch.IChHandle, inputParams map[string]interface{}) *ClientConn {
	b := &ClientConn{}
	b.Conf = clientConf
	b.Socket = *NewSocketConn(parent, handle, inputParams)
	return b
}

func (clientConn *ClientConn) GetId() string {
	return clientConn.Conf.GetAddrStr()
}

func (clientConn *ClientConn) GetChannel() gch.IChannel {
	return clientConn.Channel
}

func (clientConn *ClientConn) GetConf() IClientConf {
	return clientConn.Conf
}

func (clientConn *ClientConn) Dial() error {
	network := clientConn.GetConf().GetNetwork()
	switch network {
	case gch.NETWORK_WS:
		return dialWs(clientConn)
	case gch.NETWORK_KCP:
		return dialKcp(clientConn)
	case gch.NETWORK_TCP:
		return dialTcp(clientConn)
	case gch.NETWORK_HTTP:
		// TODO 默认为ws
		return dialWs(clientConn)
	default:
		return nil
	}
	return errors.New("start clentstrap error.")
}

// dialWs 拨号实现ws
func dialWs(clientConn *ClientConn) error {
	wsClientConf := clientConn.GetConf().(IWsClientConf)
	handle := clientConn.GetChHandle().(*gch.ChHandle)

	url := wsClientConf.GetUrl()
	params := clientConn.GetInputParams()
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
	logx.Info("dial ws url:", url)
	conn, response, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logx.Error("dial ws error:", err)
		return err
	}

	// TODO 处理resonse？
	logx.Info("ws response:", response)
	// 创建channel
	wsCh := tcpx.NewWsChannel(clientConn, conn, wsClientConf, handle, params, false)
	wsCh.SetRelativePath(wsClientConf.GetReqPath())
	err = wsCh.Start()
	if err == nil {
		clientConn.Channel = wsCh
		clientConn.Closed = false
	}
	return err
}

func dialTcp(clientConn *ClientConn) error {
	tcpClientConf := clientConn.GetConf()
	chHandle := clientConn.GetChHandle().(*gch.ChHandle)
	addr := tcpClientConf.GetAddrStr()
	logx.Info("dial tcp addr:", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logx.Error("dial tcp addr error:", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logx.Error("dial tcp conn error:", err)
		return err
	}

	tcpCh := tcpx.NewTcpChannel(clientConn, conn, tcpClientConf, chHandle, false)
	var path string
	params := clientConn.GetInputParams()
	if params != nil {
		p := params["path"]
		if p != nil {
			path = p.(string)
		}
	}
	tcpCh.SetRelativePath(path)
	err = tcpCh.Start()
	if err == nil {
		clientConn.Channel = tcpCh
		clientConn.Closed = false
	}
	return err
}

func dialKcp(clientConn *ClientConn) error {
	kcpClientConf := clientConn.GetConf()
	chHandle := clientConn.GetChHandle().(*gch.ChHandle)
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kcp addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcp conn error:", err)
		return err
	}
	kcpCh := kcpx.NewKcpChannel(clientConn, conn, kcpClientConf, chHandle, false)
	var path string
	params := clientConn.GetInputParams()
	if params != nil {
		p := params["path"]
		if p != nil {
			path = p.(string)
		}
	}
	kcpCh.SetRelativePath(path)
	err = kcpCh.Start()
	if err == nil {
		clientConn.Channel = kcpCh
		clientConn.Closed = false
	}
	return err
}

func (clientConn *ClientConn) Close() {
	if !clientConn.Closed {
		id := clientConn.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop client, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop client, id:", id)
		clientConn.Closed = true
		clientConn.Exit <- true
		clientConn.Channel.Stop()
		clientConn.Channel = nil
	}
}
