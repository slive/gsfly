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

	// GetInitTime 初始化时间
	GetInitTime() time.Time

	common.IAttact
}

// Packet channel通用packet
type Packet struct {
	channel  IChannel
	network  Network
	data     []byte
	initTime time.Time
	common.Attact
}

// GetChannel 获取对应的channel
func (b *Packet) GetChannel() IChannel {
	return b.channel
}

// Release 释放资源
func (b *Packet) Release() {
	b.data = nil
}

// IsRelease 是否以释放
func (b *Packet) IsRelease() bool {
	return b.data == nil
}

// GetNetwork 获取网络协议类型
func (b *Packet) GetNetwork() Network {
	return b.network
}

// SetNetwork 设置网络协议类型
func (b *Packet) SetNetwork(ptype Network) {
	b.network = ptype
}

// GetData 获取收发数据
func (b *Packet) GetData() []byte {
	return b.data
}

// SetData 设置收发数据
func (b *Packet) SetData(data []byte) {
	b.data = data
}

// IsPrepare 数据是否已准备好
func (b *Packet) IsPrepare() bool {
	return len(b.data) > 0
}

// GetInitTime 获取初始化的时间
func (b *Packet) GetInitTime() time.Time {
	return b.initTime
}

// NewPacket 根据不同类型创建不同的packet
// network 网络类型
func NewPacket(channel IChannel, network Network) *Packet {
	b := &Packet{
		channel:  channel,
		network:  network,
		initTime: time.Now(),
	}
	b.Attact = *common.NewAttact()
	return b
}
