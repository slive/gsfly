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

type IBootStrap interface {
	common.IParent

	// GetId 通道Id
	common.IId

	Start() error

	Stop()

	IsClosed() bool

	GetChHandle() *gch.ChannelHandle
}

type BootStrap struct {
	Closed        bool
	Exit          chan bool
	ChannelHandle *gch.ChannelHandle

	// 父接口
	*common.Parent

	*common.Id
}

func NewBootStrap(parent interface{}, handle *gch.ChannelHandle) *BootStrap {
	b := &BootStrap{
		Closed:        true,
		Exit:          make(chan bool, 1),
		ChannelHandle: handle}
	b.Parent = common.NewParent(parent)
	return b
}

func (bc *BootStrap) IsClosed() bool {
	return bc.Closed
}

func (bc *BootStrap) GetChHandle() *gch.ChannelHandle {
	return bc.ChannelHandle
}

type IServerStrap interface {
	IBootStrap
	GetConf() IServerConf
	GetChannels() map[string]gch.IChannel
}

type ServerStrap struct {
	BootStrap
	Conf     IServerConf
	Channels map[string]gch.IChannel
}

func NewServerStrap(parent interface{}, serverConf IServerConf, handle *gch.ChannelHandle) *ServerStrap {
	b := &ServerStrap{
		Conf:     serverConf,
		Channels: make(map[string]gch.IChannel),
	}
	b.BootStrap = *NewBootStrap(parent, handle)
	// OnStopHandle重新b包装
	if handle != nil {
		handle.OnStopHandle = ConverOnStopHandle(b.Channels, handle.OnStopHandle)
	}
	return b
}

func (tcpls *ServerStrap) GetId() string {
	return tcpls.Conf.GetAddrStr()
}

func (tcpls *ServerStrap) Stop() {
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
		for _, ch := range acceptChannels {
			ch.Stop()
			// delete(acceptChannels, key)
		}
	}
}

func (tcpls *ServerStrap) GetChannels() map[string]gch.IChannel {
	return tcpls.Channels
}

func (tcpls *ServerStrap) GetConf() IServerConf {
	return tcpls.Conf
}

// ConverOnStopHandle 转化OnStopHandle方法
func ConverOnStopHandle(channels map[string]gch.IChannel, onStopHandle gch.OnStopHandle) func(channel gch.IChannel) error {
	return func(channel gch.IChannel) error {
		// 释放现有资源
		delete(channels, channel.GetId())
		if onStopHandle != nil {
			return onStopHandle(channel)
		}
		return nil
	}
}

type ClientStrap struct {
	BootStrap
	Conf    IClientConf
	Channel gch.IChannel
}

type IClientStrap interface {
	IBootStrap
	GetChannel() gch.IChannel
	GetConf() IClientConf
}

func (bc *ClientStrap) GetId() string {
	return bc.Conf.GetAddrStr()
}

func (bc *ClientStrap) GetChannel() gch.IChannel {
	return bc.Channel
}

func (bc *ClientStrap) GetConf() IClientConf {
	return bc.Conf
}

func (bc *ClientStrap) Stop() {
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
