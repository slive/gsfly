/*
 * Author:slive
 * DATE:2020/8/10
 */
package kcpx

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	"gsfly/common"
	logx "gsfly/logger"
)

type Kws00Channel struct {
	KcpChannel
	onKwsMsgHandle gch.OnMsgHandle
	params         map[string]interface{}
}

// NewKws00Channel 新建KWS00 channel
// 需要实现onKwsMsgHandle 和注册（握手）成功后的onRegisterhandle
// 根据需要实现onUnRegisterhandle方法和其他ChannelHandle里的其他方法
func NewKws00Channel(parent interface{}, kcpConn *kcp.UDPSession, chConf gch.IChannelConf,
	chHandle *gch.ChannelHandle, params map[string]interface{}) *Kws00Channel {
	channel := &Kws00Channel{}
	channel.KcpChannel = *NewKcpChannel(parent, kcpConn, chConf, chHandle)
	channel.protocol = gch.PROTOCOL_KWS00
	channel.params = params
	// 更新内部kwsmsg
	chh := channel.GetChHandle()
	channel.onKwsMsgHandle = chh.GetOnMsgHandle()
	gch.UpdateMsgHandle(onInnerKws00MsgHandle, chh)
	return channel
}

// NewKws00Handle 根据需要实现onUnRegisterhandle方法和其他ChannelHandle里的其他方法
func NewKws00Handle(onKws00MsgHandle gch.OnMsgHandle, onRegisterhandle gch.OnRegisteredHandle, onUnRegisterhandle gch.OnUnRegisteredHandle) *gch.ChannelHandle {
	if onKws00MsgHandle == nil {
		errMsg := "onKws00MsgHandle is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	if onRegisterhandle == nil {
		errMsg := "onRegisterhandle is nil."
		logx.Panic(errMsg)
		panic(errMsg)
	}

	chHandle := gch.NewDefChHandle(onKws00MsgHandle)
	chHandle.OnRegisteredHandle = onRegisterhandle
	chHandle.OnUnRegisteredHandle = onUnRegisterhandle
	return chHandle
}

func (b *Kws00Channel) Start() error {
	return b.StartChannel(b)
}

func (b *Kws00Channel) Stop() {
	b.StopChannel(b)
}

func (b *Kws00Channel) Read() (gch.IPacket, error) {
	return Read(b)
}

func (b *Kws00Channel) Write(datapack gch.IPacket) error {
	return b.KcpChannel.Write(datapack)
}

func (b *Kws00Channel) GetParams() map[string]interface{} {
	return b.params
}

func (b *Kws00Channel) NewPacket() gch.IPacket {
	k := &KWS00Packet{}
	k.Packet = *gch.NewPacket(b, gch.PROTOCOL_KWS00)
	return k

}

type KWS00Packet struct {
	KcpPacket
	Frame Frame
}

func (packet *KWS00Packet) SetData(data []byte) {
	packet.Packet.SetData(data)
	packet.Frame = NewInputFrame(data)
}

// type OnKws00MsgHandle func(channel gch.IChannel, frame Frame) error
const KCP_FRAME_KEY = "kcpframe"

func onInnerKws00MsgHandle(packet gch.IPacket) error {
	srcCh := packet.GetChannel()
	defer func() {
		ret := recover()
		if ret != nil {
			logx.Error("handle error:", ret)
			srcCh.GetChHandle().OnErrorHandle(srcCh, common.NewError2(gch.ERR_REG, fmt.Sprintf("%v", ret)))
			srcCh.Stop()
		}
	}()
	protocol := srcCh.GetConf().GetProtocol()
	if protocol == gch.PROTOCOL_KWS00 {
		// 强制转换处理
		kwsPacket, ok := packet.(*KWS00Packet)
		if ok {
			frame := kwsPacket.Frame
			packet.AddAttach(KCP_FRAME_KEY, frame)
			if frame != nil {
				// 第一次建立会话时进行处理
				opCode := frame.GetOpCode()
				logx.Infof("opCode:%v, registered:%v", opCode, srcCh.IsRegistered())
				if opCode == OPCODE_TEXT_SESSION {
					registerHandle := srcCh.GetChHandle().OnRegisteredHandle
					if registerHandle != nil {
						// 注册事件
						err := registerHandle(srcCh, packet)
						if err != nil {
							logx.Error("register error:", err)
							panic("register error")
						} else {
							srcCh.SetRegistered(true)
							frame = NewOutputFrame(OPCODE_TEXT_SESSION, frame.GetPayload())
							packet.AddAttach(KCP_FRAME_KEY, frame)
						}
					}
				} else if opCode == OPCODE_CLOSE {
					unregisterHandle := srcCh.GetChHandle().OnUnRegisteredHandle
					if unregisterHandle != nil {
						// 注销事件
						unregisterHandle(srcCh, packet)
						srcCh.SetRegistered(false)
					}
				}

				kcp := srcCh.(*Kws00Channel)
				onKwsMsgHandle := kcp.onKwsMsgHandle
				if onKwsMsgHandle != nil {
					onKwsMsgHandle(packet)
				}
			} else {
				logx.Warn("frame is nil")
			}
			return nil
		}
		return nil
	}
	logx.Warn("unknown potocol:", protocol)
	return nil
}
