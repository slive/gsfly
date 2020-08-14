/*
 * channel统一的收发包
 * Author:slive
 * DATE:2020/7/21
 */
package channel

import "time"

// 定义协议类型
type Protocol int

const (
	// 协议类型
	PROTOCOL_TCP   Protocol = 1
	PROTOCOL_HTTP  Protocol = 2
	PROTOCOL_WS    Protocol = 4
	PROTOCOL_UDP   Protocol = 8
	PROTOCOL_KCP   Protocol = 16
	PROTOCOL_KWS00 Protocol = 32
	PROTOCOL_KWS01 Protocol = 64
	PROTOCOL_HTTPX Protocol = PROTOCOL_HTTP | PROTOCOL_WS
)

// IPacket 协议包接口
type IPacket interface {
	// GetChannel 获取包所属的通道
	GetChannel() IChannel

	// IsPrepare 是否准备好可以进行收发后续处理
	IsPrepare() bool

	// GetPType 协议类型
	GetPType() Protocol

	// SetPType 设置协议类型
	SetPType(ptype Protocol)

	// GetData 获取收发数据
	GetData() []byte

	// SetData 设置收发数据
	SetData(data []byte)

	GetInitTime() time.Time
}

type Packet struct {
	channel  IChannel
	ptype    Protocol
	data     []byte
	initTime time.Time
}

func (b *Packet) GetChannel() IChannel {
	return b.channel
}

func (b *Packet) GetPType() Protocol {
	return b.ptype
}

func (b *Packet) SetPType(ptype Protocol) {
	b.ptype = ptype
}

func (b *Packet) GetData() []byte {
	return b.data
}

func (b *Packet) SetData(data []byte) {
	b.data = data
}

func (b *Packet) IsPrepare() bool {
	return len(b.data) > 0
}

func (b *Packet) GetInitTime() time.Time {
	return b.initTime
}

func NewPacket(channel IChannel, ptype Protocol) *Packet {
	b := &Packet{
		channel:  channel,
		ptype:    ptype,
		initTime: time.Now(),
	}
	return b
}
