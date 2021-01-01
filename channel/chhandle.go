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

	// GetChannel 获取通道
	GetChannel() IChannel

	// GetPacket 获取收发包，可能为空
	GetPacket() IPacket

	// GetError 获取错误码
	GetError() common.GError

	// SetError 设置错误码
	SetError(gError common.GError)

	// GetRet 返回结果
	GetRet() interface{}

	// SetRet 设置返回值
	SetRet(ret interface{})

	common.IRunContext
}

// ChHandleContext channel处理handler上下文实现
type ChHandleContext struct {
	channel IChannel
	packet  IPacket
	common.Attact
	gerr common.GError
	ret  interface{}
	common.RunContext
}

// NewChHandleContext 创建channelhandle上下文
func NewChHandleContext(channel IChannel, packet IPacket) *ChHandleContext {
	c := &ChHandleContext{channel: channel, packet: packet, Attact: *common.NewAttact()}
	if packet != nil {
		c.RunContext = *common.NewRunContext(packet.GetContext())
	} else {
		c.RunContext = *common.NewRunContext(channel.GetContext())
	}
	return c
}

// GetChannel 获取通道
func (ctx *ChHandleContext) GetChannel() IChannel {
	return ctx.channel
}

// GetPacket 获取收发包
func (ctx *ChHandleContext) GetPacket() IPacket {
	return ctx.packet
}

// SetRet 设置返回值
func (ctx *ChHandleContext) SetRet(ret interface{}) {
	ctx.ret = ret
}

// GetRet 返回结果
func (ctx *ChHandleContext) GetRet() interface{} {
	return ctx.ret
}

// SetError 设置错误码
func (ctx *ChHandleContext) SetError(error common.GError) {
	ctx.gerr = error
}

// GetError 获取错误码
func (ctx *ChHandleContext) GetError() common.GError {
	return ctx.gerr
}

// ChHandleFunc channel(通信通道)处理方法
type ChHandleFunc func(ctx IChHandleContext)

// innerErrorHandle 内部错误处理，空实现
func innerErrorHandle(ctx IChHandleContext) {
	logx.ErrorTracef(ctx, "channel error, error:%v", ctx.GetError())
}

// IChHandle channel(通信通道)处理方法集接口
type IChHandle interface {
	// GetOnRead 获取读信息后的处理方法
	GetOnRead() ChHandleFunc
	// SetOnRead 设置读信息后的处理方法
	SetOnRead(onRead ChHandleFunc)

	// GetOnWrite 获取写之前的处理方法
	GetOnWrite() ChHandleFunc
	// SetOnWrite 设置写之前的处理方法
	SetOnWrite(onWrite ChHandleFunc)

	// GetOnActive 获取激活后的处理方法
	GetOnActive() ChHandleFunc
	// SetOnActive 设置激活后的处理方法
	SetOnActive(onActive ChHandleFunc)

	// GetOnInActive 获取非激活后的处理方法
	GetOnInActive() ChHandleFunc
	// SetOnInActive 设置非激活后的处理方法
	SetOnInActive(onInActive ChHandleFunc)

	// GetOnError 获取错误时方法
	GetOnError() ChHandleFunc
	// SetOnError 设置错误后的处理方法
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

// SetOnRead 设置读到数据后处理方法
func (c *ChHandle) SetOnRead(onRead ChHandleFunc) {
	c.onRead = onRead
}

// GetOnRead 获取读到数据后处理方法
func (c *ChHandle) GetOnRead() ChHandleFunc {
	return c.onRead
}

// SetOnWrite 设置写之前的处理方法
func (c *ChHandle) SetOnWrite(onWrite ChHandleFunc) {
	c.onWrite = onWrite
}

// GetOnWrite 获取写之前的处理方法
func (c *ChHandle) GetOnWrite() ChHandleFunc {
	return c.onWrite
}

// SetOnActive 设置激活后处理方法
func (c *ChHandle) SetOnActive(onActive ChHandleFunc) {
	c.onActive = onActive
}

// GetOnActive 获取激活后处理方法
func (c *ChHandle) GetOnActive() ChHandleFunc {
	return c.onActive
}

// SetOnInActive 设置非激活后处理方法
func (c *ChHandle) SetOnInActive(onInActive ChHandleFunc) {
	c.onInActive = onInActive
}

// GetOnInActive 获取非激活后处理方法
func (c *ChHandle) GetOnInActive() ChHandleFunc {
	return c.onInActive
}

// SetOnError 设置错误后处理方法
func (c *ChHandle) SetOnError(onError ChHandleFunc) {
	if onError == nil {
		c.onError = innerErrorHandle
	} else {
		c.onError = onError
	}
}

// GetOnError 获取错误后处理方法
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
		logx.PanicTracef(ctx, "implement me")
	}
}

func CopyChHandle(handle IChHandle) *ChHandle {
	newHandle := NewDefChHandle(handle.GetOnRead())
	newHandle.SetOnActive(handle.GetOnActive())
	newHandle.SetOnInActive(handle.GetOnInActive())
	newHandle.SetOnError(handle.GetOnError())
	newHandle.SetOnWrite(handle.GetOnWrite())
	return newHandle
}
