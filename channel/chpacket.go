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
type Network string

const (
	// 协议类型，移位增长
	NETWORK_TCP  Network = "tcp"
	NETWORK_HTTP Network = "http"
	NETWORK_WS   Network = "ws"
	NETWORK_UDP  Network = "udp"
	NETWORK_KCP  Network = "kcp"
)

// String 获取协议对应的字符串
func (p Network) String() string {
	switch p {
	case NETWORK_TCP:
		return "tcp"
	case NETWORK_HTTP:
		return "http"
	case NETWORK_WS:
		return "ws"
	case NETWORK_UDP:
		return "udp"
	case NETWORK_KCP:
		return "kcp"
	default:
		return "unknown"
	}
}

// IPacket 协议包接口
type IPacket interface {
	// GetChannel 获取包所属的通道
	GetChannel() IChannel

	// IsPrepare 是否准备好可以进行收发后续处理
	IsPrepare() bool

	// 释放资源
	Release()

	// 是否释放
	IsRelease() bool

	// GetNetwork 获取网络协议类型
	GetNetwork() Network

	// SetNetwork 设置网络协议类型
	SetNetwork(ptype Network)

	// GetData 获取收发数据
	GetData() []byte

	// SetData 设置收发数据
	SetData(data []byte)

	GetInitTime() time.Time

	common.IAttact
}

type Packet struct {
	channel  IChannel
	network  Network
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

func (b *Packet) IsRelease() bool {
	return b.data == nil
}

func (b *Packet) GetNetwork() Network {
	return b.network
}

func (b *Packet) SetNetwork(ptype Network) {
	b.network = ptype
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

func NewPacket(channel IChannel, network Network) *Packet {
	b := &Packet{
		channel:  channel,
		network:  network,
		initTime: time.Now(),
	}
	b.Attact = *common.NewAttact()
	return b
}
