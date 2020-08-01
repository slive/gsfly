/*
 * Author:slive
 * DATE:2020/7/31
 */
package bootstrap

import (
	gch "gsfly/channel"
	logx "gsfly/logger"
)

type Server interface {
	GetId() string

	Start() error

	Stop()

	IsClosed() bool

	GetChannelHandle() *gch.ChannelHandle
}

type BaseServer struct {
	ServerConf    *ServerConf
	Channels      map[string]gch.Channel
	Closed        bool
	Exit          chan bool
	ChannelHandle *gch.ChannelHandle
}

func (tcpls *BaseServer) GetId() string {
	return tcpls.ServerConf.GetAddrStr()
}

func (tcpls *BaseServer) IsClosed() bool {
	return tcpls.Closed
}

// GetChannelHandle 暂不支持
func (tcpls *BaseServer) GetChannelHandle() *gch.ChannelHandle {
	return tcpls.ChannelHandle
}

func (tcpls *BaseServer) Stop() {
	if !tcpls.Closed {
		tcpls.Closed = true
		tcpls.Exit <- true
		acceptChannels := tcpls.Channels
		for key, ch := range acceptChannels {
			ch.StopChannel(ch)
			delete(acceptChannels, key)
		}
		logx.Info("stop httpx listen.")
	}
}
