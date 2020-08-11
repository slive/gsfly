/*
 * Author:slive
 * DATE:2020/8/7
 */
package channel

import (
	"gsfly/common"
	logx "gsfly/logger"
)

// OnMsgHandle 处理消息方法
// packet 接收到的包，不可以为nil
type OnMsgHandle func(packet Packet) error

// OnBefWriteHandle 发送消息之前
// packet 接收到的包，不可以为nil
type OnBefWriteHandle func(packet Packet, attach ...interface{}) error

// OnAftWriteHandle 发送消息之后
// packet 接收到的包，不可以为nil
type OnAftWriteHandle func(packet Packet, attach ...interface{}) error

// OnStopHandle 处理停止时的方法
// channel 通信通道
type OnStopHandle func(channel Channel) error

// OnStartHandle 处理启动时的方法
// channel 通信通道
type OnStartHandle func(channel Channel) error

// OnRegisterHandle 处理启动时的方法
// channel 通信通道
// packet 接收到的包，可能nil
// attach 附带参数，可能为nil
type OnRegisterHandle func(channel Channel, packet Packet, attach ...interface{}) error

// OnUnRegisterHandle 处理取消注册的方法
// channel 通信通道
// packet 接收到的包，可能nil
// attach 附带参数，可能为nil
type OnUnRegisterHandle func(channel Channel, packet Packet, attach ...interface{}) error

// OnErrorHandle 处理取消注册的方法
// channel 通信通道
// err 各种异常
type OnErrorHandle func(channel Channel, gerr common.GError)

// 空实现
func innerErrorHandle(channel Channel, gerr common.GError) {
	logx.Errorf("channel error, chId:%v, error:%v", channel.GetId(), gerr)
}

type ChannelHandleMgr interface {
	SetOnStartHandle(handle OnStartHandle)
	SetOnStopHandle(handle OnStopHandle)
	SetOnRegisterHandle(handle OnRegisterHandle)
	SetOnUnRegisterHandle(handle OnUnRegisterHandle)
	SetOnErrorHandle(handle OnErrorHandle)
	SetOnAftWriteHandle(handle OnAftWriteHandle)
	SetOnBefWriteHandle(handle OnBefWriteHandle)
}

// BaseChHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	OnStartHandle      OnStartHandle
	OnRegisterHandle   OnRegisterHandle
	OnUnRegisterHandle OnUnRegisterHandle
	OnStopHandle       OnStopHandle
	OnErrorHandle      OnErrorHandle
	OnAftWriteHandle   OnAftWriteHandle
	OnBefWriteHandle   OnBefWriteHandle

	// 必须方法
	msgHandleFunc      OnMsgHandle
	innerMsgHandleFunc OnMsgHandle
}

func (c *ChannelHandle) SetOnStartHandle(handle OnStartHandle) {
	c.OnStartHandle = handle
}

func (c *ChannelHandle) SetOnStopHandle(handle OnStopHandle) {
	c.OnStopHandle = handle
}

func (c *ChannelHandle) SetOnRegisterHandle(handle OnRegisterHandle) {
	c.OnRegisterHandle = handle
}

func (c *ChannelHandle) SetOnUnRegisterHandle(handle OnUnRegisterHandle) {
	c.OnUnRegisterHandle = handle
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

// NewDefaultChHandle 创建默认的， 必须实现OnMsgHandleFunc 方法
func NewDefaultChHandle(msgHandleFunc OnMsgHandle) *ChannelHandle {
	if msgHandleFunc == nil {
		errMsg := "OnMsgHandle is nil."
		logx.Error(errMsg)
		panic(errMsg)
	}
	c := &ChannelHandle{msgHandleFunc: msgHandleFunc}
	c.innerMsgHandleFunc = c.onMsgHandle
	c.OnErrorHandle = innerErrorHandle
	return c
}

// 内部代理调用 OnMsgHandle
func (c *ChannelHandle) onMsgHandle(packet Packet) error {
	statis := packet.GetChannel().GetChStatis().HandleMsgStatics
	handleFunc := c.msgHandleFunc
	if handleFunc != nil {
		err := handleFunc(packet)
		if err != nil {
			handleStatis(statis, packet, false)
		} else {
			handleStatis(statis, packet, true)
		}
		return err
	} else {
		handleStatis(statis, packet, false)
		panic("implement me")
	}
}