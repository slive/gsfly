/*
 * Author:slive
 * DATE:2020/7/31
 */
package bootstrap

import (
	gch "github.com/Slive/gsfly/channel"
	"github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
	"github.com/emirpasic/gods/maps/hashmap"
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
	GetChannels() *hashmap.Map
}

type ServerStrap struct {
	BootStrap
	Conf     IServerConf
	Channels *hashmap.Map
}

func NewServerStrap(parent interface{}, serverConf IServerConf, chHandle *gch.ChannelHandle) *ServerStrap {
	if chHandle == nil {
		errMsg := "chHandle is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	if serverConf == nil {
		errMsg := "serverConf is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	b := &ServerStrap{
		Conf:     serverConf,
		Channels: hashmap.New(),
	}
	b.BootStrap = *NewBootStrap(parent, chHandle)
	// OnStopHandle重新b包装
	chHandle.OnStopHandle = ConverOnStopHandle(b.Channels, chHandle.OnStopHandle)
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
		acceptChannels := tcpls.Channels.Values()
		for _, ch := range acceptChannels {
			ch.(gch.IChannel).Stop()
		}
	}
}

func (tcpls *ServerStrap) GetChannels() *hashmap.Map {
	return tcpls.Channels
}

func (tcpls *ServerStrap) GetConf() IServerConf {
	return tcpls.Conf
}

// ConverOnStopHandle 转化OnStopHandle方法
func ConverOnStopHandle(channels *hashmap.Map, onStopHandle gch.OnStopHandle) func(channel gch.IChannel) error {
	return func(channel gch.IChannel) error {
		// 释放现有资源
		chId := channel.GetId()
		channels.Remove(chId)
		logx.Infof("remove serverchannel, chId:%v, channelSize:%v", chId, channels.Size())
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
