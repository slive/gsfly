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

	GetChHandle() gch.IChHandle

	GetParams() map[string]interface{}
}

type Socket struct {
	// 父接口
	cmm.Parent
	cmm.Attact
	cmm.Id
	Closed        bool
	channelHandle gch.IChHandle
	Exit          chan bool
	params        map[string] interface{}
}

func NewSocketConn(parent interface{}, handle gch.IChHandle, params map[string] interface{}) *Socket {
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

func (socketConn *Socket) GetChHandle() gch.IChHandle {
	return socketConn.channelHandle
}

func (socketConn *Socket) GetParams() map[string]interface{} {
	return socketConn.params
}
