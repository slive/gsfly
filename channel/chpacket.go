/*
 * channel统一的收发包
 * Author:slive
 * DATE:2020/7/21
 */
package channel

import (
	"github.com/Slive/gsfly/common"
	"time"
)

// 定义协议类型
type Protocol int

// String 获取协议对应的字符串
func (p Protocol) String() string {
	switch p {
	case PROTOCOL_TCP:
		return "tcp"
	case PROTOCOL_HTTP:
		return "http"
	case PROTOCOL_WS:
		return "ws"
	case PROTOCOL_UDP:
		return "udp"
	case PROTOCOL_KCP:
		return "kcp"
	case PROTOCOL_KWS00:
		return "kws00"
	case PROTOCOL_KWS01:
		return "kws001"
	case PROTOCOL_HTTPX:
		return "httpx"
	default:
		return "unknown"
	}
}

const (
	// 协议类型，移位增长
	PROTOCOL_TCP Protocol = 1 << iota
	PROTOCOL_HTTP
	PROTOCOL_WS
	PROTOCOL_UDP
	PROTOCOL_KCP
	PROTOCOL_KWS00
	PROTOCOL_KWS01
	PROTOCOL_HTTPX = PROTOCOL_HTTP | PROTOCOL_WS
)

// IPacket 协议包接口
type IPacket interface {
	// GetChannel 获取包所属的通道
	GetChannel() IChannel

	// IsPrepare 是否准备好可以进行收发后续处理
	IsPrepare() bool

	// 释放资源
	Release()

	// GetPType 协议类型
	GetPType() Protocol

	// SetPType 设置协议类型
	SetPType(ptype Protocol)

	// GetData 获取收发数据
	GetData() []byte

	// SetData 设置收发数据
	SetData(data []byte)

	GetInitTime() time.Time

	common.IAttact
}

type Packet struct {
	channel  IChannel
	ptype    Protocol
	data     []byte
	initTime time.Time
	common.Attact
}

func (b *Packet) GetChannel() IChannel {
	return b.channel
}

func (b *Packet) Release() {
	b.data = nil
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
	b.Attact = *common.NewAttact()
	return b
}
