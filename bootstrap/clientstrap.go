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

type IWsClientStrap interface {
	IClientStrap

	GetParams() map[string]interface{}
}

type WsClientStrap struct {
	ClientStrap
	params map[string]interface{}
}

func NewWsClientStrap(parent interface{}, wsClientConf IWsClientConf, handle *gch.ChannelHandle, params map[string]interface{}) IClientStrap {
	b := &WsClientStrap{}
	b.params = params
	b.Conf = wsClientConf
	b.BootStrap = *NewBootStrap(parent, handle)
	return b
}

func (wc *WsClientStrap) Start() error {
	wsClientConf := wc.GetConf().(IWsClientConf)
	handle := wc.ChannelHandle
	url := wsClientConf.GetUrl()
	params := wc.params
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
	wsCh := httpx.NewWsChannel(wc, conn, wsClientConf, handle, nil)
	err = wsCh.Start()
	if err == nil {
		wc.Channel = wsCh
		wc.Closed = false
	}
	return err
}

func (wc *WsClientStrap) GetParams() map[string]interface{} {
	return wc.params
}

type KcpClientStrap struct {
	ClientStrap
}

func NewKcpClientStrap(parent interface{}, kcpClientConf IKcpClientConf, handle *gch.ChannelHandle) IClientStrap {
	b := &KcpClientStrap{}
	b.BootStrap = *NewBootStrap(parent, handle)
	b.Conf = kcpClientConf
	return b
}

func (kc *KcpClientStrap) Start() error {
	kcpClientConf := kc.GetConf()
	chHandle := kc.ChannelHandle
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kcp addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcp conn error:", nil)
		return err
	}
	kcpCh := kcpx.NewKcpChannel(kc, conn, kcpClientConf, chHandle)
	err = kcpCh.Start()
	if err == nil {
		kc.Channel = kcpCh
		kc.Closed = false
	}
	return err
}

// Kws00ClientStrap
type Kws00ClientStrap struct {
	KcpClientStrap
	params map[string]interface{}
}

// NewKws00ClientStrap 实现kws
// onKwsMsgHandle和onRegisterhandle 必须实现，其他方法可选
func NewKws00ClientStrap(parent interface{}, kws00ClientConf IKws00ClientConf, onKwsMsgHandle gch.OnMsgHandle,
	onRegisterhandle gch.OnRegisteredHandle, onUnRegisterhandle gch.OnUnRegisteredHandle, params map[string]interface{}) IClientStrap {
	b := &Kws00ClientStrap{}
	b.Conf = kws00ClientConf
	handle := kcpx.NewKws00Handle(onKwsMsgHandle, onRegisterhandle, onUnRegisterhandle)
	b.BootStrap = *NewBootStrap(parent, handle)
	b.params = params
	return b
}

func (kc *Kws00ClientStrap) Start() error {
	clientConf := kc.GetConf()
	chHandle := kc.ChannelHandle
	addr := clientConf.GetAddrStr()
	logx.Info("dial kws00 addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kws00 conn error:", nil)
		return err
	}

	kwsCh := kcpx.NewKws00Channel(kc, conn, clientConf, chHandle, kc.params)
	err = kwsCh.Start()
	if err != nil {
		return err
	}

	// 握手操作
	if kc.params != nil{
		err = kws00Handshake(clientConf.(IKws00ClientConf), kwsCh)
	}
	if err != nil {
		kc.Stop()
	} else {
		kc.Channel = kwsCh
		kc.Closed = false
	}
	return err
}

const Kws00_Path_Key = "path"

// kws00Handshake 建立握手操作
func kws00Handshake(kcpClientConf IKws00ClientConf, kwsCh *kcpx.Kws00Channel) error {
	sessionParams := make(map[string]interface{})
	path := kcpClientConf.GetPath()
	if len(path) == 0 {
		path = ""
	}
	// 封装kws00建立连接时所需的参数
	sessionParams[Kws00_Path_Key] = path
	params := kwsCh.GetParams()
	if params != nil {
		for key, val := range params {
			sessionParams[key] = val
		}
	}

	payloadData, _ := json.Marshal(sessionParams)
	sessionFrame := kcpx.NewOutputFrame(kcpx.OPCODE_TEXT_SESSION, payloadData)
	data := sessionFrame.GetKcpData()
	logx.Info("kws00Handshake:", data)

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
	ClientStrap
}

func NewUdpClientStrap(parent interface{}, clientConf IUdpClientConf, handle *gch.ChannelHandle) IClientStrap {
	b := &UdpClientStrap{}
	b.BootStrap = *NewBootStrap(parent, handle)
	b.Conf = clientConf
	return b
}

func (uc *UdpClientStrap) Start() error {
	clientConf := uc.GetConf()
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

	udpCh := udpx.NewUdpChannel(uc, conn, clientConf, chHandle)
	err = udpCh.Start()
	if err == nil {
		uc.Channel = udpCh
		uc.Closed = false
	}
	return err
}
