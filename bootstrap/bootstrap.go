/*
 * Author:slive
 * DATE:2020/7/31
 */
package bootstrap

import (
	gch "github.com/Slive/gsfly/channel"
	cmm "github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
	"github.com/emirpasic/gods/maps/hashmap"
)

type IBootStrap interface {
	cmm.IParent

	// GetId 通道Id
	cmm.IId

	Start() error

	Stop()

	IsClosed() bool

	GetChHandle() *gch.ChannelHandle
}

type BootStrap struct {
	// 父接口
	cmm.Parent
	cmm.Id
	Closed        bool
	Exit          chan bool
	ChannelHandle *gch.ChannelHandle
}

func NewBootStrap(parent interface{}, handle *gch.ChannelHandle) *BootStrap {
	b := &BootStrap{
		Closed:        true,
		Exit:          make(chan bool, 1),
		ChannelHandle: handle}
	b.Parent = *cmm.NewParent(parent)
	return b
}

func (bs *BootStrap) IsClosed() bool {
	return bs.Closed
}

func (bs *BootStrap) GetChHandle() *gch.ChannelHandle {
	return bs.ChannelHandle
}

type IServerStrap interface {
	IBootStrap
	GetConf() IServerConf
	GetChannelPool() *hashmap.Map
}

type ServerStrap struct {
	BootStrap
	Conf        IServerConf
	ChannelPool *hashmap.Map
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
		Conf:        serverConf,
		ChannelPool: hashmap.New(),
	}
	b.BootStrap = *NewBootStrap(parent, chHandle)
	// OnStopHandle重新b包装
	chHandle.OnStopHandle = ConverOnStopHandle(b.ChannelPool, chHandle.OnStopHandle)
	return b
}

func (ss *ServerStrap) GetId() string {
	return ss.Conf.GetAddrStr()
}

func (ss *ServerStrap) Stop() {
	if !ss.Closed {
		id := ss.GetId()
		defer func() {
			ret := recover()
			logx.Infof("finish to stop listen, id:%v, ret:%v", id, ret)
		}()
		logx.Info("start to stop listen, id:", id)
		ss.Closed = true
		ss.Exit <- true
		acceptChannels := ss.GetChannelPool().Values()
		for _, ch := range acceptChannels {
			ch.(gch.IChannel).Stop()
		}
	}
}

func (ss *ServerStrap) GetChannelPool() *hashmap.Map {
	return ss.ChannelPool
}

func (ss *ServerStrap) GetConf() IServerConf {
	return ss.Conf
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
	Conf IClientConf
	// 每一个客户端只有一个channel
	Channel gch.IChannel
}

type IClientStrap interface {
	IBootStrap
	GetConf() IClientConf
	GetChannel() gch.IChannel
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
