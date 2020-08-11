/*
 * 通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package channel

import (
	"fmt"
	"github.com/pkg/errors"
	"gsfly/common"
	logx "gsfly/logger"
	"net"
)

const (
	ERR_READ  = "ERR_READ"
	ERR_MSG   = "ERR_MSG"
	ERR_WRITE = "ERR_WRITE"
	ERR_STOP  = "ERR_STOP"
	ERR_START = "ERR_START"
	ERR_REG   = "ERR_REG"
)

// Channel 通信通道接口
type Channel interface {

	// 启动通信通道
	Start() error

	// 停止通道
	Stop()

	// 通过conn写
	WriteByConn(packet Packet) error

	// 读出错是是否继续
	IsReadLoopContinued(err error) bool

	// NewPacket 创建收发包
	NewPacket() Packet

	// IsClosed 是否是关闭的
	IsClosed() bool

	// Read 读取方法
	Read() (Packet, error)

	// Write 写入方法
	Write(packet Packet) error

	// GetChConf 获取通道配置
	GetChConf() ChannelConf

	// GetChStatis 获取通道统计相关
	GetChStatis() *ChannelStatis

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	GetConn() net.Conn

	GetChHandle() *ChannelHandle

	IsRegistered() bool

	SetRegisterd(register bool)

	common.IAttact

	common.IParent

	// GetId 通道Id
	common.IId
}

// BaseChannel channel基类
type BaseChannel struct {
	ChannelHandle *ChannelHandle
	ChannelStatis *ChannelStatis
	Conn          net.Conn
	chConf        ChannelConf
	readPool      *ReadPool
	closed        bool
	closeExit     chan bool
	registered    bool

	*common.Attact
	// 父接口
	*common.Parent

	*common.Id
}

var Default_Read_Pool_Conf *ReadPoolConf

// 初始化读协程池，全局配置
var Default_ReadPool *ReadPool

func init() {
	LoadDefaultConf()
	Default_Read_Pool_Conf = Global_Conf.ReadPoolConf
	Default_ReadPool = NewReadPool(Default_Read_Pool_Conf.MaxReadPoolSize, Default_Read_Pool_Conf.MaxReadQueueSize)
}

// NewDefaultBaseChannel 创建默认基础通信通道
func NewDefaultBaseChannel(parent interface{}, chConf ChannelConf, chHandle *ChannelHandle) *BaseChannel {
	return NewBaseChannel(parent, chConf, Default_ReadPool, chHandle)
}

func NewBaseChannel(parent interface{}, chConf ChannelConf, readPool *ReadPool, chHandle *ChannelHandle) *BaseChannel {
	if chHandle == nil {
		errMsg := "ChannelHandle is nil"
		logx.Error(errMsg)
		panic(errMsg)
	}
	channel := &BaseChannel{
		ChannelHandle: chHandle,
		ChannelStatis: NewChStatis(),
		chConf:        chConf,
		readPool:      readPool,
		closeExit:     make(chan bool, 1),
	}
	channel.SetClosed(true)
	channel.SetRegisterd(false)
	channel.Attact = common.NewAttact()
	channel.Id = common.NewId()
	channel.Parent = common.NewParent(parent)
	logx.Info("create base channel, chConf:", chConf)
	return channel
}

func (b *BaseChannel) Start() error {
	return b.StartChannel(b)
}

func (b *BaseChannel) Stop() {
	b.StopChannel(b)
}

func (b *BaseChannel) StartChannel(channel Channel) error {
	id := b.GetId()
	if !channel.IsClosed() {
		return errors.New("channel is open, chId:" + id)
	}

	handle := channel.GetChHandle()
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Errorf("Start channel error, chId:%v, ret:%v", id, rec)
			err, ok := rec.(error)
			if ok {
				// 捕获处理消息异常
				handle.OnErrorHandle(channel, common.NewError1(ERR_START, err))
			}
			channel.Stop()
		}
	}()
	go b.startReadLoop(channel)

	// 启动后处理方法
	startFunc := handle.OnStartHandle
	if startFunc != nil {
		err := startFunc(channel)
		if err != nil {
			return err
		}
	}
	b.SetClosed(false)
	logx.Info("finish to start channel, chId:", channel.GetId())
	return nil
}

func (b *BaseChannel) NewPacket() Packet {
	panic("implement me")
}

func (b *BaseChannel) IsClosed() bool {
	return b.closed
}

func (b *BaseChannel) SetClosed(closed bool) {
	b.closed = closed
}

func (b *BaseChannel) Read() (packet Packet, err error) {
	panic("implement me")
}

func (b *BaseChannel) Write(datapacket Packet) error {
	if b.IsClosed() {
		return errors.New("wschannel had closed, chId:" + b.GetId())
	}

	channel := datapacket.GetChannel()
	chHandle := channel.GetChHandle()
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Error("write ws error, chId:%v, error:%v", b.GetId(), rec)
			err, ok := rec.(error)
			if !ok {
				err = errors.New(fmt.Sprintf("%v", rec))
			}
			// 捕获处理消息异常
			chHandle.OnErrorHandle(channel, common.NewError1(ERR_WRITE, err))
			// 有异常，终止执行
			channel.Stop()
		}
	}()

	if datapacket.IsPrepare() {
		// 发送前的处理
		befWriteHandle := chHandle.OnBefWriteHandle
		if befWriteHandle != nil {
			err := befWriteHandle(datapacket)
			if err != nil {
				logx.Error("befWriteHandle error:", err)
				return err
			}
		}

		// 发送
		err := channel.WriteByConn(datapacket)
		if err != nil {
			return err
		}

		SendStatis(datapacket, true)
		logx.Info(b.GetChStatis().StringSend())
		// 发送成功后的处理
		aftWriteHandle := chHandle.OnAftWriteHandle
		if aftWriteHandle != nil {
			aftWriteHandle(datapacket)
		}
		return err
	} else {
		logx.Warn("datapacket is not prepare.")
	}
	return nil
}

// WriteByConn 实现通过conn发送
func (b *BaseChannel) WriteByConn(datapacket Packet) error {
	panic("implement me")
}

func (b *BaseChannel) GetChConf() ChannelConf {
	return b.chConf
}

func (b *BaseChannel) GetChStatis() *ChannelStatis {
	return b.ChannelStatis
}

func (b *BaseChannel) GetConn() net.Conn {
	return b.Conn
}

func (b *BaseChannel) IsReadLoopContinued(err error) bool {
	return true
}

func (b *BaseChannel) GetChHandle() *ChannelHandle {
	return b.ChannelHandle
}

func (b *BaseChannel) IsRegistered() bool {
	return b.registered
}

func (b *BaseChannel) SetRegisterd(register bool) {
	b.registered = register
}

func (b *BaseChannel) StopChannel(channel Channel) {
	// 关闭状态不再执行后面的内容
	id := b.GetId()
	if channel.IsClosed() {
		logx.Info("channel is closed, chId:", id)
		return
	}

	handle := channel.GetChHandle()
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Warn("close error, chId:", id, rec)
			err, ok := rec.(error)
			if ok {
				// 捕获处理消息异常
				handle.OnErrorHandle(channel, common.NewError1(ERR_STOP, err))
			}
		} else {
			// 执行关闭后的方法
			closeFunc := handle.OnStopHandle
			if closeFunc != nil {
				closeFunc(channel)
			}
			logx.Info("finish to close channel, chId:", id)
		}
	}()

	logx.Info("start to close channel, chId:", id)
	// 清理关闭相关
	b.SetClosed(true)
	b.closeExit <- true
	close(b.closeExit)
	conn := channel.GetConn()
	if conn != nil {
		conn.Close()
	}
}

// StartReadLoop 启动循环读取，读取到数据包后，放入#ReadQueue中，等待处理
func (b *BaseChannel) startReadLoop(channel Channel) {
	defer func() {
		rec := recover()
		if rec != nil {
			err, ok := rec.(error)
			if ok {
				logx.Errorf("readloop error, chId:%v, err:%v", b.GetId(), err)
				// 捕获处理消息异常
				channel.GetChHandle().OnErrorHandle(channel, common.NewError1(ERR_READ, err))
				b.Stop()
			}
		}
	}()
	logx.Info("start to readloop, chId:", b.GetId())
	for {
		select {
		case <-b.closeExit:
			logx.Info("stop read loop.")
			return
		default:
			rev, err := channel.Read()
			if err != nil {
				if !channel.IsReadLoopContinued(err) {
					logx.Panic("read loop error:", err)
					return
				} else {
					continue
				}
			}

			if rev != nil && rev.IsPrepare() {
				readPool := b.readPool
				if readPool != nil {
					// 放入读取协程池等待处理
					readPool.Cache(rev)
				} else {
					// 否则默认直接处理
					channel.GetChHandle().innerMsgHandleFunc(rev)
				}
			}}
	}
}

func (b *BaseChannel) LocalAddr() net.Addr {
	return nil
}

func (b *BaseChannel) RemoteAddr() net.Addr {
	return nil
}
