/*
 * 通用的socket处理类
 * Author:slive
 * DATE:2020/7/31
 */
package socket

import (
	gch "github.com/Slive/gsfly/channel"
	cmm "github.com/Slive/gsfly/common"
)

// ISocket socket接口
type ISocket interface {
	cmm.IParent

	cmm.IAttact

	cmm.IId

	Close()

	IsClosed() bool

	GetChHandle() gch.IChHandle

	// GetInputParams 建立socketconn所需的参数
	GetInputParams() map[string]interface{}

	cmm.IRunContext
}

// Socket socketconn通信
type Socket struct {
	// 父接口
	cmm.Parent
	cmm.Attact
	cmm.Id
	Closed        bool
	channelHandle gch.IChHandle
	Exit          chan bool
	params        map[string]interface{}

	cmm.RunContext
}

// NewSocket 创建socketconn
// parent 父类
// chHandle handle
// inputParams 所需参数
func NewSocket(parent interface{}, chHandle gch.IChHandle, inputParams map[string]interface{}) *Socket {
	b := &Socket{
		Closed:        true,
		Exit:          make(chan bool, 1),
		channelHandle: chHandle}
	b.Parent = *cmm.NewParent(parent)
	b.Attact = *cmm.NewAttact()
	b.params = inputParams
	b.RunContext = *cmm.NewRunContextByParent(parent)
	return b
}

func (socket *Socket) SetId(id string) {
	socket.AddTrace(id)
	socket.Id.SetId(id)
}

func (socket *Socket) IsClosed() bool {
	return socket.Closed
}

func (socket *Socket) GetChHandle() gch.IChHandle {
	return socket.channelHandle
}

// GetInputParams 建立socketconn所需的参数
func (socket *Socket) GetInputParams() map[string]interface{} {
	return socket.params
}
