/*
 * Author:slive
 * DATE:2020/7/28
 */
package bootstrap

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	"gsfly/channel/tcp/ws"
	"gsfly/channel/udp"
	kcpx "gsfly/channel/udp/kcp"
	logx "gsfly/logger"
	"net"
)

type WsClient struct {
	BaseClient
	ClientConf *WsClientConf
}

func NewWsClient(wsClientConf *WsClientConf, handle *gch.ChannelHandle) Client {
	b := &WsClient{
		ClientConf: wsClientConf,
	}
	b.BaseCommunication = *NewCommunication(handle)
	return b
}

func (wc *WsClient) Start() error {
	wsClientConf := wc.ClientConf
	handle := wc.ChannelHandle
	url := wsClientConf.GetUrl()
	params := wsClientConf.Params
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
	wsCh := ws.NewWsChannelWithHandle(conn, &wsClientConf.BaseChannelConf, handle)
	err = wsCh.StartChannel(wsCh)
	if err == nil {
		wc.Channel = wsCh
	}
	return err
}

type KcpClient struct {
	BaseClient
	ClientConf *KcpClientConf
}

func NewKcpClient(kcpClientConf *KcpClientConf, handle *gch.ChannelHandle) Client {
	b := &KcpClient{
		ClientConf: kcpClientConf,
	}
	b.BaseCommunication = *NewCommunication(handle)
	return b
}

func (kc *KcpClient) Start() error {
	kcpClientConf := kc.ClientConf
	chHandle := kc.ChannelHandle
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return err
	}
	kcpCh := kcpx.NewKcpChannelWithHandle(conn, &kcpClientConf.BaseChannelConf, chHandle)
	err = kcpCh.StartChannel(kcpCh)
	if err == nil {
		kc.Channel = kcpCh
	}
	return err
}

type UdpClient struct {
	BaseClient
	ClientConf *UdpClientConf
}

func NewUdpClient(clientConf *UdpClientConf, handle *gch.ChannelHandle) Client {
	b := &UdpClient{
		ClientConf: clientConf,
	}
	b.BaseCommunication = *NewCommunication(handle)
	return b
}

func (uc *UdpClient) Start() error {
	clientConf := uc.ClientConf
	chHandle := uc.ChannelHandle
	addr := clientConf.GetAddrStr()
	logx.Info("dial udp addr:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.Error("resolve updaddr error:", err)
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logx.Error("dial udp conn error:", nil)
		return err
	}

	udpCh := udp.NewUdpChannelWithHandle(conn, &clientConf.BaseChannelConf, chHandle)
	err = udpCh.StartChannel(udpCh)
	if err == nil {
		uc.Channel = udpCh
	}
	return err
}
