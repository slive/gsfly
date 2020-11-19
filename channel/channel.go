/*
 * 通用的通信通道，主要是负责通信通道的收发+事件的处理
 * Author:slive
 * DATE:2020/7/17
 */
package channel

import (
	"fmt"
	"github.com/Slive/gsfly/common"
	logx "github.com/Slive/gsfly/logger"
	"github.com/pkg/errors"
	"io"
	"net"
	"sync"
)

const (
	// 错误相关的定义
	ERR_READ     = "ERR_READ"
	ERR_MSG      = "ERR_MSG"
	ERR_WRITE    = "ERR_WRITE"
	ERR_INACTIVE = "ERR_INACTIVE"
	ERR_ACTIVE   = "ERR_ACTIVE"
	ERR_REG      = "ERR_REG"
)

// IChannel 通信通道接口
type IChannel interface {

	// Start 启动通信通道
	Start() error

	// Stop 停止通道
	Stop()

	// WriteByConn 通过conn写
	WriteByConn(packet IPacket) error

	// IsReadLoopContinued 读出错是是否继续
	IsReadLoopContinued(err error) bool

	// NewPacket 创建收发包
	NewPacket() IPacket

	// IsClosed 是否是关闭的
	IsClosed() bool

	// Read 读取方法
	Read() (IPacket, error)

	// Write 写入方法
	Write(packet IPacket) error

	// GetConf 获取通道配置
	GetConf() IChannelConf

	// GetChStatis 获取通道统计相关
	GetChStatis() *ChannelStatis

	// LocalAddr 本地地址
	LocalAddr() net.Addr

	// RemoteAddr 远程地址
	RemoteAddr() net.Addr

	// GetConn 获取原始的conn
	GetConn() net.Conn

	// GetChHandle 获取handle相关类
	GetChHandle() *ChHandle

	// IsActived 是否是激活状态
	IsActived() bool

	// SetActived 设置是否激活
	SetActived(register bool)

	// IsServer 是否是服务端产生的channel
	IsServer() bool

	// GetReqPath 获取path路径
	GetRelativePath() string

	// SetRelativePath 设置path
	SetRelativePath(path string)

	common.IAttact

	common.IParent

	// GetId 通道Id
	common.IId
}

// Channel channel基类
type Channel struct {
	ChannelHandle *ChHandle
	ChannelStatis *ChannelStatis
	Conn          net.Conn
	conf          IChannelConf
	readPool      *ReadPool
	closed        bool
	closeExit     chan bool
	actived       bool
	server        bool

	// 路径，根据各自需要定义
	relativePath string

	// 父接口
	common.Parent
	common.Id
	common.Attact

	readBuf []byte
}

// 初始化读协程池，全局配置
var def_readPoolConf *ReadPoolConf

// 默认读线程池
var def_readPool *ReadPool

// 默认channel配置
var def_channel_Conf IChannelConf

func initDefChannelConfs() {
	if def_readPool == nil {
		def_readPoolConf = Global_Conf.ReadPoolConf
		def_readPool = NewReadPool(def_readPoolConf.MaxReadPoolSize, def_readPoolConf.MaxReadQueueSize)
	}

	if def_channel_Conf == nil {
		def_channel_Conf = Global_Conf.ChannelConf
	}

	logx.Info("init default readPoolConf:", def_readPoolConf)
	logx.Info("init default channelConf:", def_channel_Conf)
}

// InitChannelConfs 初始化channel相关配置，如果该方法未调用，则调用默认初始化方法
// readPoolConf 读线程池配置
// chConf channel相关配置
func InitChannelConfs(readPoolConf *ReadPoolConf, chConf *ChannelConf) {
	if readPoolConf == nil {
		err := "ReadPoolConf is nil"
		logx.Panic(err)
		panic(err)
	}

	if chConf == nil {
		err := "ChannelConf is nil"
		logx.Panic(err)
		panic(err)
	}

	def_readPool = NewReadPool(readPoolConf.MaxReadPoolSize, readPoolConf.MaxReadQueueSize)
	def_channel_Conf = chConf
}

var initOnce sync.Once

// NewDefChannel 创建默认基础通信通道
func NewDefChannel(parent interface{}, chConf IChannelConf, chHandle *ChHandle, server bool) *Channel {
	// 全局初始化一次
	return NewChannel(parent, chConf, def_readPool, chHandle, server)
}

// NewSimpleChannel 创建默认基础通信通道
// onReadHandler 消息处理方法
func NewSimpleChannel(onReadHandler ChHandleFunc) *Channel {
	// 全局初始化一次
	chHandle := NewDefChHandle(onReadHandler)
	return NewChannel(nil, nil, nil, chHandle, false)
}

// NewChannel 创建channel
// parent 父节点，可为nil
// chConf channel配置，可为nil，如果为nil，则选用默认
// readPool 读取消息池，可为nil，如果为nil，则选用默认
// chHandle 处理handle，包括读写，注册等处理，不可为空
// server 是否为服务端创建
func NewChannel(parent interface{}, chConf IChannelConf, readPool *ReadPool, chHandle *ChHandle, server bool) *Channel {
	if chHandle == nil {
		errMsg := "ChHandle is nil"
		logx.Error(errMsg)
		panic(errMsg)
	}

	// 如果未初始化一些必要配置，则默认初始化
	initOnce.Do(func() {
		// 默认初始化logger
		logx.InitDefLogger()
		// 默认初始化channel
		initDefChannelConfs()
	})

	// 选用默认
	if chConf == nil {
		chConf = def_channel_Conf
	}
	if readPool == nil {
		readPool = def_readPool
	}

	channel := &Channel{
		ChannelHandle: chHandle,
		ChannelStatis: NewChStatis(),
		conf:          chConf,
		readPool:      readPool,
		closeExit:     make(chan bool, 1),
		server:        server,
		relativePath:  "",
	}

	channel.SetClosed(true)
	channel.SetActived(false)
	channel.Attact = *common.NewAttact()
	channel.Id = *common.NewId()
	channel.Parent = *common.NewParent(parent)
	logx.Info("create base channel, chConf:%+v", chConf)
	channel.readBuf = make([]byte, chConf.GetReadBufSize())
	return channel
}

func (ch *Channel) Start() error {
	return ch.StartChannel(ch)
}

func (ch *Channel) Stop() {
	ch.StopChannel(ch)
}

func (ch *Channel) StartChannel(channel IChannel) error {
	id := ch.GetId()
	if !channel.IsClosed() {
		return errors.New("channel is open, chId:" + id)
	}

	ctx := NewChHandleContext(channel, nil)
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Errorf("connect channel error, chId:%v, ret:%v", id, rec)
			err, ok := rec.(error)
			if ok {
				// 捕获处理消息异常
				NotifyErrorHandle(ctx, err, ERR_ACTIVE)
			}
			channel.Stop()
		}
	}()
	go ch.startReadLoop(channel)

	ch.SetClosed(false)
	logx.Info("finish to start channel, chId:", channel.GetId())
	return nil
}

func (ch *Channel) GetReadBuf() []byte {
	return ch.readBuf
}

func NotifyErrorHandle(ctx IChHandleContext, err error, errMsg string) {
	chHandle := ctx.GetChannel().GetChHandle()
	errorHandler := chHandle.GetOnError()
	ctx.SetError(common.NewError1(errMsg, err))
	errorHandler(ctx)
}

func (ch *Channel) NewPacket() IPacket {
	panic("implement me")
}

func (ch *Channel) IsClosed() bool {
	return ch.closed
}

func (ch *Channel) SetClosed(closed bool) {
	ch.closed = closed
}

func (ch *Channel) Read() (packet IPacket, err error) {
	panic("implement me")
}

func (ch *Channel) Write(datapacket IPacket) error {
	if ch.IsClosed() {
		return errors.New("wschannel had closed, chId:" + ch.GetId())
	}

	channel := datapacket.GetChannel()
	ctx := NewChHandleContext(channel, datapacket)
	chHandle := channel.GetChHandle()
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Errorf("write ws error, chId:%v, error:%v", ch.GetId(), rec)
			err, ok := rec.(error)
			if !ok {
				err = errors.New(fmt.Sprintf("%v", rec))
			}
			// 捕获处理消息异常
			NotifyErrorHandle(ctx, err, ERR_WRITE)
			// 有异常，终止执行
			channel.Stop()
		}
	}()

	if datapacket.IsPrepare() {
		// 发送前的处理
		onWriteHandle := chHandle.onWrite
		if onWriteHandle != nil {
			onWriteHandle(ctx)
			err := ctx.gerr
			if err != nil {
				logx.Error("onWriteHandle error:", err)
				return err
			}
		}

		// 发送
		err := channel.WriteByConn(datapacket)
		if err != nil {
			return err
		}

		SendStatis(datapacket, true)
		return err
	} else {
		logx.Warn("datapacket is not prepare.")
	}
	return nil
}

// WriteByConn 实现通过conn发送
func (ch *Channel) WriteByConn(datapacket IPacket) error {
	panic("implement me")
}

func (ch *Channel) GetConf() IChannelConf {
	return ch.conf
}

func (ch *Channel) GetChStatis() *ChannelStatis {
	return ch.ChannelStatis
}

func (ch *Channel) GetConn() net.Conn {
	return ch.Conn
}

func (ch *Channel) IsReadLoopContinued(err error) bool {
	// 读取超过一定失败次数后，不再继续执行
	return ch.GetChStatis().RevStatics.FailTimes < (int64)(ch.conf.GetCloseRevFailTime())
}

func (ch *Channel) GetChHandle() *ChHandle {
	return ch.ChannelHandle
}

func (ch *Channel) IsActived() bool {
	return ch.actived
}

func (ch *Channel) SetActived(actived bool) {
	ch.actived = actived
}

// 是否是服务端产生的channel
func (ch *Channel) IsServer() bool {
	return ch.server
}

// GetReqPath 获取相对path，根据各自业务需要进行定义
func (ch *Channel) GetRelativePath() string {
	return ch.relativePath
}

// SetRelativePath 设置相对path，根据各自业务需要进行定义
func (ch *Channel) SetRelativePath(relativePath string) {
	ch.relativePath = relativePath
}

func (ch *Channel) StopChannel(channel IChannel) {
	// 关闭状态不再执行后面的内容
	id := ch.GetId()
	if channel.IsClosed() {
		logx.Info("channel is closed, chId:", id)
		return
	}

	ctx := NewChHandleContext(channel, nil)
	handle := channel.GetChHandle()
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Warnf("close error, chId:%v, err:%v", id, rec)
			err, ok := rec.(error)
			if ok {
				// 捕获处理消息异常
				NotifyErrorHandle(ctx, err, ERR_INACTIVE)
			}
		}

		// 执行关闭后的方法
		closeFunc := handle.onInActive
		if closeFunc != nil {
			closeFunc(ctx)
		}
		logx.Info("finish to close channel, chId:", id)

	}()

	logx.Info("start to close channel, chId:", id)
	// 清理关闭相关
	ch.SetClosed(true)
	ch.Clear()
	ch.closeExit <- true
	close(ch.closeExit)

	// TODO udpchannel没必要关闭，待定，关闭conn不应该channel来管理？
	conn := channel.GetConn()
	if conn != nil && NETWORK_UDP != ch.GetConf().GetNetwork() {
		conn.Close()
	}
}

// StartReadLoop 启动循环读取，读取到数据包后，放入#ReadQueue中，等待处理
func (ch *Channel) startReadLoop(channel IChannel) {
	chId := ch.GetId()
	ctx := NewChHandleContext(channel, nil)
	defer func() {
		rec := recover()
		if rec != nil {
			logx.Errorf("readloop error, chId:%v, err:%v", chId, rec)
			err, ok := rec.(error)
			if ok {
				// 捕获处理消息异常
				NotifyErrorHandle(ctx, err, ERR_READ)
			}
			ch.Stop()
		}
	}()
	logx.Info("start to readloop, chId:", chId)
	for {
		select {
		case <-ch.closeExit:
			logx.Info("stop read loop, chId:", chId)
			return
		default:
			rev, err := channel.Read()
			if err != nil {
				switch err {
				case io.EOF, io.ErrClosedPipe, io.ErrUnexpectedEOF:
					// io的异常直接结束
					logx.Panicf("readloop io error:%v", err)
				default:
					// 其他异常循环等待或者忽略
					if !channel.IsReadLoopContinued(err) {
						logx.Panicf("readloop error:%v", err)
					} else {
						continue
					}
				}
			}

			if rev != nil && rev.IsPrepare() {
				readPool := ch.readPool
				if readPool != nil {
					// 放入读取协程池等待处理
					readPool.Cache(rev)
				} else {
					// 否则默认直接处理
					context := NewChHandleContext(channel, rev)
					channel.GetChHandle().onInnerRead(context)
				}
			}}
	}
}

func HandleOnActive(ctx IChHandleContext) {
	channel := ctx.GetChannel()
	activeFunc := channel.GetChHandle().GetOnActive()
	if activeFunc != nil {
		activeFunc(ctx)
		gerr := ctx.GetError()
		if gerr != nil {
			NotifyErrorHandle(ctx, gerr.GetErr(), ERR_READ)
		} else {
			channel.SetActived(true)
		}
	}
}

func (ch *Channel) LocalAddr() net.Addr {
	return nil
}

func (ch *Channel) RemoteAddr() net.Addr {
	return nil
}
