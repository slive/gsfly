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

func DialWs(wsClientConf *WsClientConf, handle *gch.ChannelHandle) (gch.Channel, error) {
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
		return nil, err
	}

	// TODO 处理resonse？
	logx.Info("ws response:", response)
	wsCh := ws.NewWsChannelWithHandle(conn, &wsClientConf.ChannelConf, handle)
	err = wsCh.StartChannel(wsCh)
	return wsCh, err
}

func DialKcp(kcpClientConf *KcpClientConf, chHandle *gch.ChannelHandle, protocol gch.Protocol) (gch.Channel, error) {
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return nil, err
	}
	kcpCh := kcpx.NewKcpChannelWithHandle(conn, &kcpClientConf.ChannelConf, chHandle, protocol)
	err = kcpCh.StartChannel(kcpCh)
	return kcpCh, err
}

func DialUdp(clientConf *UdpClientConf, chHandle *gch.ChannelHandle) (gch.Channel, error) {
	addr := clientConf.GetAddrStr()
	logx.Info("dial udp addr:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.Error("resolve updaddr error:", err)
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logx.Error("dial udp conn error:", nil)
		return nil, err
	}
	udpCh := udp.NewUdpChannelWithHandle(conn, &clientConf.ChannelConf, chHandle)
	err = udpCh.StartChannel(udpCh)
	return udpCh, err
}
