/*
 * Author:slive
 * DATE:2020/7/31
 */
package socket

import (
	gch "github.com/Slive/gsfly/channel"
	cmm "github.com/Slive/gsfly/common"
)

type ISocket interface {
	cmm.IParent

	cmm.IAttact

	cmm.IId

	Close()

	IsClosed() bool

	GetChHandle() gch.IChannelHandle

	GetParams() []interface{}
}

type Socket struct {
	// 父接口
	cmm.Parent
	cmm.Attact
	cmm.Id
	Closed        bool
	channelHandle gch.IChannelHandle
	Exit          chan bool
	params        []interface{}
}

func NewSocketConn(parent interface{}, handle gch.IChannelHandle, params ...interface{}) *Socket {
	b := &Socket{
		Closed:        true,
		Exit:          make(chan bool, 1),
		channelHandle: handle}
	b.Parent = *cmm.NewParent(parent)
	b.Attact = *cmm.NewAttact()
	b.params = params
	return b
}

func (socketConn *Socket) IsClosed() bool {
	return socketConn.Closed
}

func (socketConn *Socket) GetChHandle() gch.IChannelHandle {
	return socketConn.channelHandle
}

func (socketConn *Socket) GetParams() []interface{} {
	return socketConn.params
}
