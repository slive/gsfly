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
	gch "gsfly/channel"
	"gsfly/channel/tcp/ws"
	"gsfly/config"
	logx "gsfly/logger"
	httpx "net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 15,
	ReadBufferSize:   10 * 1024,
	WriteBufferSize:  10 * 1024,
}

type HttpxListen struct {
	config.HttpxServerConf
	httpHandlers map[string]HttpHandleFunc
	wsHandlers   map[string]*gch.ChannelHandle
}

// Http和Websocket 的服务监听
func NewHttpxListen(httpxServerConf config.HttpxServerConf) *HttpxListen {
	t := &HttpxListen{
		HttpxServerConf: httpxServerConf,
		httpHandlers:    make(map[string]HttpHandleFunc),
		wsHandlers:      make(map[string]*gch.ChannelHandle),
	}
	return t
}

// AddHttpHandleFunc 添加http处理方法
func (t *HttpxListen) AddHttpHandleFunc(pattern string, httpHandleFunc HttpHandleFunc) {
	t.httpHandlers[pattern] = httpHandleFunc
}

// AddWsHandleFunc 添加Websocket处理方法
func (t *HttpxListen) AddWsHandleFunc(pattern string, wsHandleFunc *gch.ChannelHandle) {
	t.wsHandlers[pattern] = wsHandleFunc
}

func StartHttpxListen(tcpls *HttpxListen) {
	// http处理事件
	httpHandlers := tcpls.httpHandlers
	if httpHandlers != nil {
		for key, f := range httpHandlers {
			httpx.HandleFunc(key, f)
		}
	}

	// ws处理事件
	acceptChannels := make(map[string]gch.Channel, 10)
	defer func() {
		for key, ch := range acceptChannels {
			ch.StopChannel()
			delete(acceptChannels, key)
		}
		logx.Info("stop httpx listen.")
	}()

	wsHandlers := tcpls.wsHandlers
	if wsHandlers != nil {
		for key, f := range wsHandlers {
			httpx.HandleFunc(key, func(writer httpx.ResponseWriter, r *httpx.Request) {
				logx.Info("requestWs:", r.URL)
				err := startWs(writer, r, f, acceptChannels)
				if err != nil {
					logx.Error("start ws error:", err)
				}
			})
		}
	}

	addr := tcpls.GetAddrStr()
	logx.Info("start httpx listen, addr:", addr)
	httpx.ListenAndServe(addr, nil)
	select {}
}

type HttpHandleFunc func(httpx.ResponseWriter, *httpx.Request)

// startWs 启动ws处理
func startWs(w httpx.ResponseWriter, r *httpx.Request, handle *gch.ChannelHandle, acceptChannels map[string]gch.Channel) error {
	connLen := len(acceptChannels)
	maxAcceptSize := config.Global_Conf.TcpServerConf.MaxAcceptSize
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

	wsCh := ws.NewWsChannelWithHandle(conn, config.Global_Conf.ChannelConf, handle)
	err = wsCh.StartChannel(wsCh)
	if err != nil {
		acceptChannels[wsCh.GetChId()] = wsCh
	}
	return err
}

func DialWs(wsClientConf *config.WsClientConf, handle *gch.ChannelHandle) (gch.Channel, error) {
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
