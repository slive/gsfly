/*
 * Author:slive
 * DATE:2020/7/21
 */
package channel

const (
	// 协议类型，如0:tcp,1:udp,2:http,3:websocket,4:kcpws
	PROTOCOL_TCP = iota
	PROTOCOL_HTTP
	PROTOCOL_WS
	PROTOCOL_UDP
	PROTOCOL_KCP
	PROTOCOL_KWS
	PROTOCOL_KHTTP
)

type Packet interface {
	GetChannel() Channel

	GetPrepare() bool

	// GetPType 协议类型，如0:tcp,1:http,2:websocket...
	GetPType() int

	SetPType(ptype int)

	GetData() []byte

	SetData(data []byte)
}

type Basepacket struct {
	channel Channel
	ptype   int
	data    []byte
}

func (b *Basepacket) GetChannel() Channel {
	return b.channel
}

func (b *Basepacket) GetPType() int {
	return b.ptype
}

func (b *Basepacket) SetPType(ptype int) {
	b.ptype = ptype
}

func (b *Basepacket) GetData() []byte {
	return b.data
}

func (b *Basepacket) SetData(data []byte) {
	b.data = data
}

func (b *Basepacket) GetPrepare() bool {
	return len(b.data) > 0
}

func NewBasePacket(channel Channel, ptype int) *Basepacket {
	b := &Basepacket{
		channel: channel,
		ptype:   ptype,
	}
	return b
}
