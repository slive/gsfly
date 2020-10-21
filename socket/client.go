/*
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
)

type IClientConn interface {
	ISocket
	GetConf() IClientConf
	GetChannel() gch.IChannel
	Dial() error
}

type ClientConn struct {
	Socket
	Conf IClientConf
	// 每一个客户端只有一个channel
	Channel gch.IChannel
}

func NewClientConn(parent interface{}, clientConf IClientConf, handle gch.IChannelHandle, params ...interface{}) *ClientConn {
	b := &ClientConn{}
	b.Conf = clientConf
	b.Socket = *NewSocketConn(parent, handle, params...)
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
	default:
		return nil
	}
	return errors.New("start clentstrap error.")
}

func dialWs(clientStrap *ClientConn) error {
	wsClientConf := clientStrap.GetConf().(IWsClientConf)
	handle := clientStrap.GetChHandle().(*gch.ChannelHandle)
	url := wsClientConf.GetUrl()
	params := getWsParams(clientStrap)
	if params != nil && len(params) > 0 {
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
	wsCh := tcpx.NewWsChannel(clientStrap, conn, wsClientConf, handle, params, false)
	err = wsCh.Start()
	if err == nil {
		clientStrap.Channel = wsCh
		clientStrap.Closed = false
	}
	return err
}

func getWsParams(clientStrap *ClientConn) map[string]interface{} {
	params := clientStrap.GetParams()
	if len(params) > 0 {
		return params[0].(map[string]interface{})
	}
	return nil
}

func dialKcp(clientStrap *ClientConn) error {
	kcpClientConf := clientStrap.GetConf()
	chHandle := clientStrap.GetChHandle().(*gch.ChannelHandle)
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kcp addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcp conn error:", nil)
		return err
	}
	kcpCh := kcpx.NewKcpChannel(clientStrap, conn, kcpClientConf, chHandle, false)
	err = kcpCh.Start()
	if err == nil {
		clientStrap.Channel = kcpCh
		clientStrap.Closed = false
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
