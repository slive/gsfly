/*
 * channel统一的收发包
 * Author:slive
 * DATE:2020/7/21
 */
package channel

import (
	"fmt"
	"github.com/slive/gsfly/common"
	"time"
)

// 定义协议类型
type Network string

const (
	// 协议类型，移位增长
	NETWORK_TCP     Network = "tcp"
	NETWORK_HTTP    Network = "http"
	NETWORK_WS      Network = "ws"
	NETWORK_UDP     Network = "udp"
	NETWORK_KCP     Network = "kcp"
	NETWORK_UNKNOWN Network = ""
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

func ToNetwork(network string) Network {
	switch network {
	case "tcp":
		return NETWORK_TCP
	case "http":
		return NETWORK_HTTP
	case "ws":
		return NETWORK_WS
	case "udp":
		return NETWORK_UDP
	case "kcp":
		return NETWORK_KCP
	default:
		return NETWORK_UNKNOWN
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

	common.IRunContext
}

// Packet channel通用packet
type Packet struct {
	channel  IChannel
	network  Network
	data     []byte
	initTime time.Time
	common.Attact
	common.RunContext
}

// GetChannel 获取对应的channel
func (packet *Packet) GetChannel() IChannel {
	return packet.channel
}

// Release 释放资源
func (packet *Packet) Release() {
	packet.data = nil
}

// IsRelease 是否以释放
func (packet *Packet) IsRelease() bool {
	return packet.data == nil
}

// GetNetwork 获取网络协议类型
func (packet *Packet) GetNetwork() Network {
	return packet.network
}

// SetNetwork 设置网络协议类型
func (packet *Packet) SetNetwork(ptype Network) {
	packet.network = ptype
}

// GetData 获取收发数据
func (packet *Packet) GetData() []byte {
	return packet.data
}

// SetData 设置收发数据
func (packet *Packet) SetData(data []byte) {
	packet.data = data
}

// IsPrepare 数据是否已准备好
func (packet *Packet) IsPrepare() bool {
	return len(packet.data) > 0
}

// GetInitTime 获取初始化的时间
func (packet *Packet) GetInitTime() time.Time {
	return packet.initTime
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
	b.RunContext = *common.NewRunContext(channel.GetContext())
	b.AppendTrace(fmt.Sprintf("%v", time.Now().UnixNano()))
	return b
}
