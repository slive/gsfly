/*
 * Author:slive
 * DATE:2020/8/10
 */
package kcpx

import (
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	"gsfly/common"
	logx "gsfly/logger"
)

type Kws00Channel struct {
	KcpChannel
	onKwsMsgHandle OnKws00MsgHandle
}

// NewKws00Channel 新建KWS00 channel
// 需要实现onKwsMsgHandle 和注册（握手）成功后的onRegisterhandle
// 根据需要实现onUnRegisterhandle方法和其他ChannelHandle里的其他方法
func NewKws00Channel(parent interface{}, kcpConn *kcp.UDPSession, chConf gch.ChannelConf, onKwsMsgHandle OnKws00MsgHandle, chHandle *gch.ChannelHandle) *Kws00Channel {
	if onKwsMsgHandle == nil {
		logx.Panic("onKwsMsgHandle is nil.")
		return nil
	}

	channel := &Kws00Channel{}
	channel.KcpChannel = *NewKcpChannel(parent, kcpConn, chConf, chHandle)
	channel.protocol = gch.PROTOCOL_KWS00
	channel.onKwsMsgHandle = onKwsMsgHandle
	return channel
}

// NewKws00Handle 根据需要实现onUnRegisterhandle方法和其他ChannelHandle里的其他方法
func NewKws00Handle(onRegisterhandle gch.OnRegisterHandle, onUnRegisterhandle gch.OnUnRegisterHandle) *gch.ChannelHandle {
	if onRegisterhandle == nil {
		logx.Panic("onRegisterhandle is nil.")
		return nil
	}

	chHandle := gch.NewDefaultChHandle(onKws00MsgHandle)
	chHandle.OnRegisterHandle = onRegisterhandle
	chHandle.OnUnRegisterHandle = onUnRegisterhandle
	return chHandle
}

func (b *Kws00Channel) Start() error {
	return b.StartChannel(b)
}

func (b *Kws00Channel) Stop() {
	b.StopChannel(b)
}

func (b *Kws00Channel) Read() (gch.Packet, error) {
	return Read(b)
}

func (b *Kws00Channel) Write(datapack gch.Packet) error {
	return b.KcpChannel.Write(datapack)
}

func (b *Kws00Channel) NewPacket() gch.Packet {
	k := &KWS00Packet{}
	k.Basepacket = *gch.NewBasePacket(b, gch.PROTOCOL_KWS00)
	return k

}

type KWS00Packet struct {
	KcpPacket
	Frame Frame
}

func (packet *KWS00Packet) SetData(data []byte) {
	packet.Basepacket.SetData(data)
	packet.Frame = NewInputFrame(data)
}

type OnKws00MsgHandle func(channel gch.Channel, frame Frame) error

func onKws00MsgHandle(packet gch.Packet) error {
	srcCh := packet.GetChannel()
	defer func() {
		ret := recover()
		if ret != nil {
			err := ret.(error)
			logx.Error("handle error:", err)
			srcCh.GetChHandle().OnErrorHandle(srcCh, common.NewError1(gch.ERR_REG, err))
			srcCh.Stop()
		}
	}()
	protocol := srcCh.GetChConf().GetProtocol()
	if protocol == gch.PROTOCOL_KWS00 {
		// 强制转换处理
		kwsPacket, ok := packet.(*KWS00Packet)
		if ok {
			frame := kwsPacket.Frame
			if frame != nil {
				// 第一次建立会话时进行处理
				if frame.GetOpCode() == OPCODE_TEXT_SESSION {
					registerHandle := srcCh.GetChHandle().OnRegisterHandle
					if registerHandle != nil {
						// 注册事件
						err := registerHandle(srcCh, packet, frame)
						if err != nil {
							logx.Error("register error:", err)
							panic("register error")
						} else {
							srcCh.SetRegisterd(true)
						}
					}
				} else if frame.GetOpCode() == OPCODE_CLOSE {
					unregisterHandle := srcCh.GetChHandle().OnUnRegisterHandle
					if unregisterHandle != nil {
						// 注销事件
						unregisterHandle(srcCh, packet, frame)
						srcCh.SetRegisterd(false)
					}
				}

				kcp := srcCh.(*Kws00Channel)
				onKwsMsgHandle := kcp.onKwsMsgHandle
				if onKwsMsgHandle != nil {
					onKwsMsgHandle(kcp, frame)
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
