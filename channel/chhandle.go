/*
 * Author:slive
 * DATE:2020/8/7
 */
package channel

import (
	"github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
)

type IChHandlerContext interface {
	common.IAttact
	GetChannel() IChannel
	GetPacket() IPacket

	GetError() common.GError
	SetError(gError common.GError)

	GetRet() struct{}
	SetRet(ret struct{})
}

type ChHandlerContext struct {
	channel IChannel
	packet  IPacket
	common.Attact
	gerr common.GError
	ret  struct{}
}

func NewChHandlerContext(channel IChannel, packet IPacket) *ChHandlerContext {
	return &ChHandlerContext{channel: channel, packet: packet, Attact: *common.NewAttact()}
}

func (ctx *ChHandlerContext) GetChannel() IChannel {
	return ctx.channel
}

func (ctx *ChHandlerContext) GetPacket() IPacket {
	return ctx.packet
}

func (ctx *ChHandlerContext) SetRet(ret struct{}) {
	ctx.ret = ret
}

func (ctx *ChHandlerContext) GetRet() struct{} {
	return ctx.ret
}

func (ctx *ChHandlerContext) SetError(error common.GError) {
	ctx.gerr = error
}

func (ctx *ChHandlerContext) GetError() common.GError {
	return ctx.gerr
}

type ChHandler func(ctx IChHandlerContext)

// 空实现
func innerErrorHandle(ctx IChHandlerContext) {
	logx.Errorf("channel error, chId:%v, error:%v", ctx.GetChannel().GetId(), ctx.GetError())
}

type IChannelHandle interface {
	GetOnReadHandler() ChHandler
	SetOnReadHandler(onReadHandler ChHandler)

	GetOnWriteHandler() ChHandler
	SetOnWriteHandler(onWriteHandler ChHandler)

	GetOnActiveHandler() ChHandler
	SetOnActiveHandler(onActiveHandler ChHandler)

	GetOnInActiveHandler() ChHandler
	SetOnInActiveHandler(onInActiveHandler ChHandler)

	GetOnErrorHandler() ChHandler
	SetOnErrorHandler(onErrorHandler ChHandler)
}

// ChannelHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	onActiveHandler    ChHandler
	onInActiveHandler  ChHandler
	onInnerReadHandler ChHandler
	onReadHandler      ChHandler
	onWriteHandler     ChHandler
	onErrorHandler     ChHandler
}

func (c *ChannelHandle) SetOnReadHandler(onReadHandler ChHandler) {
	c.onReadHandler = onReadHandler
}

func (c *ChannelHandle) GetOnReadHandler() ChHandler {
	return c.onReadHandler
}

func (c *ChannelHandle) SetOnWriteHandler(onWriteHandler ChHandler) {
	c.onWriteHandler = onWriteHandler
}

func (c *ChannelHandle) GetOnWriteHandler() ChHandler {
	return c.onWriteHandler
}

func (c *ChannelHandle) SetOnActiveHandler(onActiveHandler ChHandler) {
	c.onActiveHandler = onActiveHandler
}

func (c *ChannelHandle) GetOnActiveHandler() ChHandler {
	return c.onActiveHandler
}

func (c *ChannelHandle) SetOnInActiveHandler(onInActiveHandler ChHandler) {
	c.onInActiveHandler = onInActiveHandler
}

func (c *ChannelHandle) GetOnInActiveHandler() ChHandler {
	return c.onInActiveHandler
}

func (c *ChannelHandle) SetOnErrorHandler(errHandler ChHandler) {
	if errHandler == nil {
		c.onErrorHandler = innerErrorHandle
	} else {
		c.onErrorHandler = errHandler
	}
}

func (c *ChannelHandle) GetOnErrorHandler() ChHandler {
	return c.onErrorHandler
}

// NewDefChHandle 创建默认的， 必须实现OnMsgHandleFunc 方法
func NewDefChHandle(onReadHandler ChHandler) *ChannelHandle {
	if onReadHandler == nil {
		errMsg := "onReadHandler is nil."
		logx.Error(errMsg)
		panic(errMsg)
	}
	c := &ChannelHandle{}
	c.SetOnReadHandler(onReadHandler)
	c.SetOnErrorHandler(innerErrorHandle)
	c.onInnerReadHandler = c.onWapperReadHandler
	return c
}

// 内部代理调用 OnMsgHandle
func (c *ChannelHandle) onWapperReadHandler(ctx IChHandlerContext) {
	handleFunc := c.GetOnReadHandler()
	packet := ctx.GetPacket()
	if handleFunc != nil {
		handleFunc(ctx)
		err := ctx.GetError()
		// 记录统计相关信息
		if err != nil {
			HandleMsgStatis(packet, false)
			errHandler := c.GetOnErrorHandler()
			errHandler(ctx)
		} else {
			HandleMsgStatis(packet, true)
		}
	} else {
		HandleMsgStatis(packet, false)
		panic("implement me")
	}
}
