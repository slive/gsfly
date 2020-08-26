/*
 * Author:slive
 * DATE:2020/8/7
 */
package channel

import (
	"github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
)

// OnMsgHandle 处理消息方法
// packet 接收到的包，不可以为nil
type OnMsgHandle func(packet IPacket) error

// OnBefWriteHandle 发送消息之前
// packet 发送的包，不可以为nil，发送前处理失败，则不继续执行
type OnBefWriteHandle func(packet IPacket) error

// OnAftWriteHandle 发送消息之后
// packet 发送的包，不可以为nil
type OnAftWriteHandle func(packet IPacket) error

// OnStopHandle 处理停止时的方法
// channel 通信通道
type OnStopHandle func(channel IChannel) error

// OnStartHandle 处理启动时的方法
// channel 通信通道， 启动有问题则不再继续执行
type OnStartHandle func(channel IChannel) error

// OnRegisteredHandle 处理注册后（可能成功也可能失败）的方法
// channel 通信通道
// packet 接收到的包，可能nil
// attach 附带参数，可能为nil
type OnRegisteredHandle func(channel IChannel, packet IPacket) error

// OnUnRegisteredHandle 处理取消注册的方法
// channel 通信通道
// packet 接收到的包，可能nil
// attach 附带参数，可能为nil
type OnUnRegisteredHandle func(channel IChannel, packet IPacket) error

// OnErrorHandle 处理取消注册的方法
// channel 通信通道
// err 各种异常
type OnErrorHandle func(channel IChannel, gerr common.GError)

// 空实现
func innerErrorHandle(channel IChannel, gerr common.GError) {
	logx.Errorf("channel error, chId:%v, error:%v", channel.GetId(), gerr)
}

type IChannelHandle interface {
	GetOnMsgHandle() OnMsgHandle

	SetOnStartHandle(handle OnStartHandle)
	SetOnStopHandle(handle OnStopHandle)
	SetOnRegisteredHandle(handle OnRegisteredHandle)
	SetOnUnRegisteredHandle(handle OnUnRegisteredHandle)
	SetOnErrorHandle(handle OnErrorHandle)
	SetOnAftWriteHandle(handle OnAftWriteHandle)
	SetOnBefWriteHandle(handle OnBefWriteHandle)
}

// ChannelHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	OnStartHandle        OnStartHandle
	OnRegisteredHandle   OnRegisteredHandle
	OnUnRegisteredHandle OnUnRegisteredHandle
	OnStopHandle         OnStopHandle
	OnErrorHandle        OnErrorHandle
	OnAftWriteHandle     OnAftWriteHandle
	OnBefWriteHandle     OnBefWriteHandle

	// 处理消息方法，必须方法
	msgHandleFunc      OnMsgHandle
	innerMsgHandleFunc OnMsgHandle
}

func (c *ChannelHandle) SetOnStartHandle(handle OnStartHandle) {
	c.OnStartHandle = handle
}

func (c *ChannelHandle) SetOnStopHandle(handle OnStopHandle) {
	c.OnStopHandle = handle
}

func (c *ChannelHandle) SetOnRegisteredHandle(handle OnRegisteredHandle) {
	c.OnRegisteredHandle = handle
}

func (c *ChannelHandle) SetOnUnRegisteredHandle(handle OnUnRegisteredHandle) {
	c.OnUnRegisteredHandle = handle
}

func (c *ChannelHandle) SetOnErrorHandle(handle OnErrorHandle) {
	if handle == nil {
		c.OnErrorHandle = innerErrorHandle
	} else {
		c.OnErrorHandle = handle
	}
}

func (c *ChannelHandle) SetOnAftWriteHandle(handle OnAftWriteHandle) {
	c.OnAftWriteHandle = handle
}

func (c *ChannelHandle) SetOnBefWriteHandle(handle OnBefWriteHandle) {
	c.OnBefWriteHandle = handle
}

func (c *ChannelHandle) GetOnMsgHandle() OnMsgHandle {
	return c.msgHandleFunc
}

// NewDefChHandle 创建默认的， 必须实现OnMsgHandleFunc 方法
func NewDefChHandle(msgHandleFunc OnMsgHandle) *ChannelHandle {
	if msgHandleFunc == nil {
		errMsg := "OnMsgHandle is nil."
		logx.Error(errMsg)
		panic(errMsg)
	}
	c := &ChannelHandle{}
	c.msgHandleFunc = msgHandleFunc
	c.innerMsgHandleFunc = c.onMsgHandle
	c.OnErrorHandle = innerErrorHandle
	return c
}

func UpdateMsgHandle(msgHandleFunc OnMsgHandle, chHandle *ChannelHandle) {
	if msgHandleFunc == nil {
		errMsg := "OnMsgHandle is nil."
		logx.Error(errMsg)
		panic(errMsg)
	}
	chHandle.msgHandleFunc = msgHandleFunc
}

// 内部代理调用 OnMsgHandle
func (c *ChannelHandle) onMsgHandle(packet IPacket) error {
	handleFunc := c.msgHandleFunc
	if handleFunc != nil {
		err := handleFunc(packet)
		// 记录统计相关信息
		if err != nil {
			HandleMsgStatis(packet, false)
		} else {
			HandleMsgStatis(packet, true)
		}
		return err
	} else {
		HandleMsgStatis(packet, false)
		panic("implement me")
	}
}

func CopyChannelHandle(srcChHandle *ChannelHandle) *ChannelHandle {
	ch := NewDefChHandle(srcChHandle.GetOnMsgHandle())
	ch.SetOnStopHandle(srcChHandle.OnStopHandle)
	ch.SetOnStartHandle(srcChHandle.OnStartHandle)
	ch.SetOnAftWriteHandle(srcChHandle.OnAftWriteHandle)
	ch.SetOnBefWriteHandle(srcChHandle.OnBefWriteHandle)
	ch.SetOnErrorHandle(srcChHandle.OnErrorHandle)
	ch.SetOnRegisteredHandle(srcChHandle.OnRegisteredHandle)
	ch.SetOnUnRegisteredHandle(srcChHandle.OnUnRegisteredHandle)
	return ch
}
