/*
 * Author:slive
 * DATE:2020/7/28
 */
package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	httpx "gsfly/channel/tcpx/httpx"
	"gsfly/channel/udpx"
	kcpx "gsfly/channel/udpx/kcpx"
	logx "gsfly/logger"
	"net"
)

type WsClientStrap struct {
	BaseClientStrap
	ClientConf *WsClientConf
}

func NewWsClient(parent interface{}, wsClientConf *WsClientConf, handle *gch.ChannelHandle) ClientStrap {
	b := &WsClientStrap{
		ClientConf: wsClientConf,
	}
	b.BaseBootStrap = *NewBaseBootStrap(parent, handle)
	return b
}

func (wc *WsClientStrap) Start() error {
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
	wsCh := httpx.NewWsChannel(wc, conn, &wsClientConf.BaseChannelConf, handle)
	err = wsCh.Start()
	if err == nil {
		wc.Channel = wsCh
	}
	return err
}

type KcpClientStrap struct {
	BaseClientStrap
	ClientConf *KcpClientConf
}

func NewKcpClient(parent interface{}, kcpClientConf *KcpClientConf, handle *gch.ChannelHandle) ClientStrap {
	b := &KcpClientStrap{
		ClientConf: kcpClientConf,
	}
	b.BaseBootStrap = *NewBaseBootStrap(parent, handle)
	return b
}

func (kc *KcpClientStrap) Start() error {
	kcpClientConf := kc.ClientConf
	chHandle := kc.ChannelHandle
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kcp addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcp conn error:", nil)
		return err
	}
	kcpCh := kcpx.NewKcpChannel(kc, conn, &kcpClientConf.BaseChannelConf, chHandle)
	err = kcpCh.Start()
	if err == nil {
		kc.Channel = kcpCh
	}
	return err
}

// Kws00ClientStrap
type Kws00ClientStrap struct {
	KcpClientStrap
	ClientConf     *Kws00ClientConf
	onKwsMsgHandle kcpx.OnKws00MsgHandle
}

// NewKws00Client 实现kws
// onKwsMsgHandle和onRegisterhandle 必须实现，其他方法可选
func NewKws00Client(parent interface{}, kws00ClientConf *Kws00ClientConf, onKwsMsgHandle kcpx.OnKws00MsgHandle,
	onRegisterhandle gch.OnRegisterHandle, onUnRegisterhandle gch.OnUnRegisterHandle) ClientStrap {
	b := &Kws00ClientStrap{}
	b.ClientConf = kws00ClientConf
	handle := kcpx.NewKws00Handle(onRegisterhandle, onUnRegisterhandle)
	b.BaseBootStrap = *NewBaseBootStrap(parent, handle)
	b.onKwsMsgHandle = onKwsMsgHandle
	return b
}

func (kc *Kws00ClientStrap) Start() error {
	kcpClientConf := kc.ClientConf
	chHandle := kc.ChannelHandle
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kws00 addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kws00 conn error:", nil)
		return err
	}

	kwsCh := kcpx.NewKws00Channel(kc, conn, &kcpClientConf.BaseChannelConf, kc.onKwsMsgHandle, chHandle)
	err = kwsCh.Start()
	if err != nil {
		return err
	}

	// 握手操作
	err = handshake(kcpClientConf, kwsCh)
	if err != nil {
		kc.Stop()
	} else {
		kc.Channel = kwsCh
	}
	return err
}

// handshake 建立握手操作
func handshake(kcpClientConf *Kws00ClientConf, kwsCh *kcpx.Kws00Channel) error {
	sessionParams := make(map[string]interface{})
	path := kcpClientConf.Path
	if len(path) == 0 {
		path = ""
	}
	sessionParams["path"] = path
	params := kcpClientConf.Params
	if params != nil {
		for key, val := range params {
			sessionParams[key] = val
		}
	}
	payloadData, _ := json.Marshal(sessionParams)
	sessionFrame := kcpx.NewOutputFrame(kcpx.OPCODE_TEXT_SESSION, payloadData)
	data := sessionFrame.GetKcpData()
	logx.Info("handshake:", data)

	packet := kwsCh.NewPacket()
	packet.SetData(data)
	err := kwsCh.Write(packet)
	if err == nil {
		logx.Info("write packet:", packet)
	} else {
		logx.Error(err)
	}
	return err
}

type UdpClientStrap struct {
	BaseClientStrap
	ClientConf *UdpClientConf
}

func NewUdpClient(parent interface{}, clientConf *UdpClientConf, handle *gch.ChannelHandle) ClientStrap {
	b := &UdpClientStrap{
		ClientConf: clientConf,
	}
	b.BaseBootStrap = *NewBaseBootStrap(parent, handle)
	return b
}

func (uc *UdpClientStrap) Start() error {
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

	udpCh := udpx.NewUdpChannel(uc, conn, &clientConf.BaseChannelConf, chHandle)
	err = udpCh.Start()
	if err == nil {
		uc.Channel = udpCh
	}
	return err
}
