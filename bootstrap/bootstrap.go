/*
 * Author:slive
 * DATE:2020/7/31
 */
package bootstrap

import (
	gch "gsfly/channel"
	logx "gsfly/logger"
)

type Communication interface {
	GetId() string

	Start() error

	Stop()

	IsClosed() bool

	GetChannelHandle() *gch.ChannelHandle
}

type BaseCommunication struct {
	Closed        bool
	Exit          chan bool
	ChannelHandle *gch.ChannelHandle
}

func NewCommunication(handle *gch.ChannelHandle) *BaseCommunication {
	b := &BaseCommunication{
		Closed:        true,
		Exit:          make(chan bool, 1),
		ChannelHandle: handle}
	return b
}

func (bc *BaseCommunication) IsClosed() bool {
	return bc.Closed
}

func (bc *BaseCommunication) GetChannelHandle() *gch.ChannelHandle {
	return bc.ChannelHandle
}

type Server interface {
	Communication
}

type BaseServer struct {
	BaseCommunication
	ServerConf ServerConf
	Channels   map[string]gch.Channel
}

func (tcpls *BaseServer) GetId() string {
	return tcpls.ServerConf.GetAddrStr()
}

func (tcpls *BaseServer) Stop() {
	if !tcpls.Closed {
		id := tcpls.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop httpx listen, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop httpx listen, id:", id)
		tcpls.Closed = true
		tcpls.Exit <- true
		acceptChannels := tcpls.Channels
		for key, ch := range acceptChannels {
			ch.StopChannel(ch)
			delete(acceptChannels, key)
		}
	}
}

type BaseClient struct {
	BaseCommunication
	ClientConf ClientConf
	Channel    gch.Channel
}

type Client interface {
	Communication
	GetChannel() gch.Channel
	GetClientConf() ClientConf
}

func (bc *BaseClient) GetId() string {
	return bc.ClientConf.GetAddrStr()
}

func (bc *BaseClient) GetChannel() gch.Channel {
	return bc.Channel
}

func (bc *BaseClient) GetClientConf() ClientConf {
	return bc.ClientConf
}

func (bc *BaseClient) Stop() {
	if !bc.Closed {
		id := bc.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop client, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop client, id:", id)
		bc.Closed = true
		bc.Exit <- true
		bc.Channel.StopChannel(bc.Channel)
		bc.Channel = nil
	}
}
