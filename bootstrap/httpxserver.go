/*
 * Author:slive
 * DATE:2020/7/17
 */
package bootstrap

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	gchannel "gsfly/channel"
	"gsfly/channel/tcp/ws"
	"gsfly/config"
	logx "gsfly/logger"
	httpx "net/http"
)

var acceptChannels = make(map[string]gchannel.Channel, 100)
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1000,
	ReadBufferSize:   10 * 1024,
	WriteBufferSize:  0,
}

type HttpxListen struct {
	ip           string
	port         int
	httpHandlers map[string]HttpHandleFunc
	wsHandlers   map[string]gchannel.ChannelHandle
}

func NewHttpxListen(ip string, port int) *HttpxListen {
	t := &HttpxListen{
		ip:           ip,
		port:         port,
		httpHandlers: make(map[string]HttpHandleFunc),
		wsHandlers:   make(map[string]gchannel.ChannelHandle),
	}
	return t
}

func (t *HttpxListen) AddHttpHandleFunc(pattern string, httpHandleFunc HttpHandleFunc) {
	t.httpHandlers[pattern] = httpHandleFunc
}

func (t *HttpxListen) AddWsHandleFunc(pattern string, wsHandleFunc gchannel.ChannelHandle) {
	t.wsHandlers[pattern] = wsHandleFunc
}

func StartListen(tcpls *HttpxListen) {
	handlers := tcpls.httpHandlers
	if handlers != nil {
		for key, f := range handlers {
			httpx.HandleFunc(key, f)
		}
	}

	wsHandlers := tcpls.wsHandlers
	if wsHandlers != nil {
		for key, f := range wsHandlers {
			httpx.HandleFunc(key, func(writer httpx.ResponseWriter, r *httpx.Request) {
				logx.Info("requestWs:", r.URL)
				err := startWs(writer, r, f)
				if err != nil {
					logx.Error("start ws error:", err)
				}
			})
		}
	}

	addr := tcpls.ip + fmt.Sprintf(":%v", tcpls.port)
	logx.Info("start httpx listen, addr:", addr)
	httpx.ListenAndServe(addr, nil)
	defer func() {
		for _, conn := range acceptChannels {
			conn.Close()
		}
		logx.Info("stop httpx listen.")
	}()
	select {}
}

type HttpHandleFunc func(httpx.ResponseWriter, *httpx.Request)

func startWs(w httpx.ResponseWriter, r *httpx.Request, handle gchannel.ChannelHandle) error {
	connLen := len(acceptChannels)
	maxAcceptSize := config.Global_Conf.TcpServerConf.MaxAcceptSize
	if connLen >= maxAcceptSize {
		return errors.New("max accept size:" + fmt.Sprintf("%v", maxAcceptSize))
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Println("upgrade error:", err)
		return err
	}
	if err != nil {
		logx.Error("accept error:", nil)
		return err
	}
	ch, err := ws.StartWsChannelWithHandle(conn, config.Global_Conf.ChannelConf, handle)
	acceptChannels[ch.GetChId()] = ch
	return err
}
