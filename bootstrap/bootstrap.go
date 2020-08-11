/*
 * Author:slive
 * DATE:2020/7/31
 */
package bootstrap

import (
	gch "gsfly/channel"
	"gsfly/common"
	logx "gsfly/logger"
)

type BootStrap interface {
	common.IParent

	// GetId 通道Id
	common.IId

	Start() error

	Stop()

	IsClosed() bool

	GetChannelHandle() *gch.ChannelHandle
}

type BaseBootStrap struct {
	Closed        bool
	Exit          chan bool
	ChannelHandle *gch.ChannelHandle

	// 父接口
	*common.Parent

	*common.Id
}

func NewBaseBootStrap(parent interface{}, handle *gch.ChannelHandle) *BaseBootStrap {
	b := &BaseBootStrap{
		Closed:        true,
		Exit:          make(chan bool, 1),
		ChannelHandle: handle}
	b.Parent = common.NewParent(parent)
	return b
}

func (bc *BaseBootStrap) IsClosed() bool {
	return bc.Closed
}

func (bc *BaseBootStrap) GetChannelHandle() *gch.ChannelHandle {
	return bc.ChannelHandle
}

type ServerStrap interface {
	BootStrap
}

type BaseServerStrap struct {
	BaseBootStrap
	ServerConf ServerConf
	Channels   map[string]gch.Channel
}

func (tcpls *BaseServerStrap) GetId() string {
	return tcpls.ServerConf.GetAddrStr()
}

func (tcpls *BaseServerStrap) Stop() {
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
			ch.Stop()
			delete(acceptChannels, key)
		}
	}
}

type BaseClientStrap struct {
	BaseBootStrap
	ClientConf ClientConf
	Channel    gch.Channel
}

type ClientStrap interface {
	BootStrap
	GetChannel() gch.Channel
	GetClientConf() ClientConf
}

func (bc *BaseClientStrap) GetId() string {
	return bc.ClientConf.GetAddrStr()
}

func (bc *BaseClientStrap) GetChannel() gch.Channel {
	return bc.Channel
}

func (bc *BaseClientStrap) GetClientConf() ClientConf {
	return bc.ClientConf
}

func (bc *BaseClientStrap) Stop() {
	if !bc.Closed {
		id := bc.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop client, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop client, id:", id)
		bc.Closed = true
		bc.Exit <- true
		bc.Channel.Stop()
		bc.Channel = nil
	}
}
