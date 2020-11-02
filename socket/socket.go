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
}

// NewSocketConn 创建socketconn
// parent 父类
// chHandle handle
// inputParams 所需参数
func NewSocketConn(parent interface{}, chHandle gch.IChHandle, inputParams map[string]interface{}) *Socket {
	b := &Socket{
		Closed:        true,
		Exit:          make(chan bool, 1),
		channelHandle: chHandle}
	b.Parent = *cmm.NewParent(parent)
	b.Attact = *cmm.NewAttact()
	b.params = inputParams
	return b
}

func (socketConn *Socket) IsClosed() bool {
	return socketConn.Closed
}

func (socketConn *Socket) GetChHandle() gch.IChHandle {
	return socketConn.channelHandle
}

// GetInputParams 建立socketconn所需的参数
func (socketConn *Socket) GetInputParams() map[string]interface{} {
	return socketConn.params
}
