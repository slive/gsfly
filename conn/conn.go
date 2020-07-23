/*
 * 通信连接
 * Author:slive
 * DATE:2020/7/17
 */
package conn

type Conn interface {
	GetChId() string

	Read(packet Packet) error

	ReadLoop() error

	Write(packet Packet) error

	SetHandleMsgFunc(handleMsgFunc HandleMsgFunc)

	GetHandleMsgFunc() HandleMsgFunc

	Close()
}

type HandleMsgFunc func(conn Conn, packet Packet) error

type BaseConn struct {
	handleMsgFunc HandleMsgFunc
	closeExit     chan bool
}

func NewBaseConn() *BaseConn {
	conn := &BaseConn{closeExit: make(chan bool, 1)}
	return conn
}

func (b *BaseConn) GetChId() string {
	panic("unsupported")
}

func (b *BaseConn) Read(datapack interface{}) error {
	panic("unsupported")
}

func (b *BaseConn) ReadLoop() error {
	panic("unsupported")
}

func (b *BaseConn) Write(datapack interface{}) (int, error) {
	panic("unsupported")
}

func (b *BaseConn) SetHandleMsgFunc(handleMsgFunc HandleMsgFunc) {
	b.handleMsgFunc = handleMsgFunc
}

func (b *BaseConn) GetHandleMsgFunc() HandleMsgFunc {
	return b.handleMsgFunc
}

func (b *BaseConn) Close() {
	b.closeExit <- true
	close(b.closeExit)
}
