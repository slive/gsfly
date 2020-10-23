/*
 * Author:slive
 * DATE:2020/8/7
 */
package channel

import (
	"github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
)

// IChHandleContext channel处理handler上下文接口
type IChHandleContext interface {
	common.IAttact

	GetChannel() IChannel

	GetPacket() IPacket

	GetError() common.GError

	SetError(gError common.GError)

	// GetRet 返回结果
	GetRet() interface{}

	SetRet(ret interface{})
}

// ChHandleContext channel处理handler上下文实现
type ChHandleContext struct {
	channel IChannel
	packet  IPacket
	common.Attact
	gerr common.GError
	ret  interface{}
}

func NewChHandleContext(channel IChannel, packet IPacket) *ChHandleContext {
	return &ChHandleContext{channel: channel, packet: packet, Attact: *common.NewAttact()}
}

func (ctx *ChHandleContext) GetChannel() IChannel {
	return ctx.channel
}

func (ctx *ChHandleContext) GetPacket() IPacket {
	return ctx.packet
}

func (ctx *ChHandleContext) SetRet(ret interface{}) {
	ctx.ret = ret
}

func (ctx *ChHandleContext) GetRet() interface{} {
	return ctx.ret
}

func (ctx *ChHandleContext) SetError(error common.GError) {
	ctx.gerr = error
}

func (ctx *ChHandleContext) GetError() common.GError {
	return ctx.gerr
}

// ChHandleFunc channel(通信通道)处理方法
type ChHandleFunc func(ctx IChHandleContext)

// 空实现
func innerErrorHandle(ctx IChHandleContext) {
	logx.Errorf("channel error, chId:%v, error:%v", ctx.GetChannel().GetId(), ctx.GetError())
}

// IChHandle channel(通信通道)处理方法集接口
type IChHandle interface {
	GetOnRead() ChHandleFunc
	SetOnRead(onRead ChHandleFunc)

	GetOnWrite() ChHandleFunc
	SetOnWrite(onWrite ChHandleFunc)

	GetOnActive() ChHandleFunc
	SetOnActive(onActive ChHandleFunc)

	GetOnInActive() ChHandleFunc
	SetOnInActive(onInActive ChHandleFunc)

	GetOnError() ChHandleFunc
	SetOnError(onError ChHandleFunc)
}

// ChHandle channel(通信通道)处理集，针对如开始，关闭和收到消息的方法
type ChHandle struct {
	onActive    ChHandleFunc
	onInActive  ChHandleFunc
	onInnerRead ChHandleFunc
	onRead      ChHandleFunc
	onWrite     ChHandleFunc
	onError     ChHandleFunc
}

func (c *ChHandle) SetOnRead(onRead ChHandleFunc) {
	c.onRead = onRead
}

func (c *ChHandle) GetOnRead() ChHandleFunc {
	return c.onRead
}

func (c *ChHandle) SetOnWrite(onWrite ChHandleFunc) {
	c.onWrite = onWrite
}

func (c *ChHandle) GetOnWrite() ChHandleFunc {
	return c.onWrite
}

func (c *ChHandle) SetOnActive(onActive ChHandleFunc) {
	c.onActive = onActive
}

func (c *ChHandle) GetOnActive() ChHandleFunc {
	return c.onActive
}

func (c *ChHandle) SetOnInActive(onInActive ChHandleFunc) {
	c.onInActive = onInActive
}

func (c *ChHandle) GetOnInActive() ChHandleFunc {
	return c.onInActive
}

func (c *ChHandle) SetOnError(onError ChHandleFunc) {
	if onError == nil {
		c.onError = innerErrorHandle
	} else {
		c.onError = onError
	}
}

func (c *ChHandle) GetOnError() ChHandleFunc {
	return c.onError
}

// NewDefChHandle 创建默认，要求必须实现onReadHandler方法
func NewDefChHandle(onReadHandler ChHandleFunc) *ChHandle {
	if onReadHandler == nil {
		errMsg := "onRead is nil."
		logx.Error(errMsg)
		panic(errMsg)
	}
	c := &ChHandle{}
	c.SetOnRead(onReadHandler)
	c.SetOnError(innerErrorHandle)
	c.onInnerRead = c.onWapperReadHandler
	return c
}

// 内部代理调用 OnMsgHandle
func (c *ChHandle) onWapperReadHandler(ctx IChHandleContext) {
	handleFunc := c.GetOnRead()
	packet := ctx.GetPacket()
	if handleFunc != nil {
		handleFunc(ctx)
		err := ctx.GetError()
		// 记录统计相关信息
		if err != nil {
			HandleMsgStatis(packet, false)
			errHandler := c.GetOnError()
			errHandler(ctx)
		} else {
			HandleMsgStatis(packet, true)
		}
	} else {
		HandleMsgStatis(packet, false)
		panic("implement me")
	}
}
