/*
 * 通信通道
 * Author:slive
 * DATE:2020/7/17
 */
package channel

import (
	"fmt"
	"github.com/pkg/errors"
	logx "gsfly/logger"
	"net"
	"time"
)

// Channel 通信通道接口
type Channel interface {

	// 启动通信通道
	StartChannel(channel Channel) error

	// NewPacket 创建收发包
	NewPacket() Packet

	// GetChId 通道Id
	GetChId() string

	// IsClosed 是否是关闭的
	IsClosed() bool

	// Read 读取方法
	Read() (Packet, error)

	// Write 写入方法
	Write(packet Packet) error

	// GetHandleMsgFunc 处理读取消息方法的方法
	GetHandleMsgFunc() HandleMsgFunc

	// GetChConf 获取通道配置
	GetChConf() *BaseChannelConf

	// GetChStatis 获取通道统计相关
	GetChStatis() *ChannelStatis

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	// 停止通道
	StopChannel(channel Channel)

	GetConn() net.Conn

	IsReadLoopContinued(err error) bool
}

// ChannelHandle 通信通道处理结构，针对如开始，关闭和收到消息的方法
type ChannelHandle struct {
	HandleStartFunc HandleStartFunc
	HandleMsgFunc   HandleMsgFunc
	HandleCloseFunc HandleCloseFunc
}

// HandleMsgFunc 处理消息方法
type HandleMsgFunc func(packet Packet) error

// HandleCloseFunc 处理关闭时的方法
type HandleCloseFunc func(channel Channel) error

// HandleStartFunc 处理启动时的方法
type HandleStartFunc func(channel Channel) error

// NewChHandle 创建处理方法类
func NewChHandle(handleMsgFunc HandleMsgFunc, handleStartFunc HandleStartFunc, handleCloseFunc HandleCloseFunc) *ChannelHandle {
	return &ChannelHandle{
		HandleStartFunc: handleStartFunc,
		HandleMsgFunc:   handleMsgFunc,
		HandleCloseFunc: handleCloseFunc,
	}
}

// ChannelStatis 统计相关，比如收发包数目，收发次数
type ChannelStatis struct {
	RevByteNum    int64
	RevPacketNum  int64
	RevTime       time.Time
	SendByteNum   int64
	SendPacketNum int64
	SendTime      time.Time
}

func NewChStatis() *ChannelStatis {
	return &ChannelStatis{
		RevByteNum:    0,
		RevPacketNum:  0,
		RevTime:       time.Now(),
		SendByteNum:   0,
		SendPacketNum: 0,
		SendTime:      time.Now(),
	}
}

func (s *ChannelStatis) StringSend() string {
	return fmt.Sprintf("SendTime:%v, SendByteNum:%v, SendPacketNum:%v", s.SendTime, s.SendByteNum, s.SendPacketNum)
}
func (s *ChannelStatis) StringRev() string {
	return fmt.Sprintf("RevTime:%v, RevByteNum:%v, RevPacketNum:%v", s.RevTime, s.RevByteNum, s.RevPacketNum)
}

// BaseChannel channel基类
type BaseChannel struct {
	ChannelHandle
	ChannelStatis
	// conn      net.Conn
	chConf    *BaseChannelConf
	readPool  *ReadPool
	chId      string
	closed    bool
	closeExit chan bool
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
func NewDefaultBaseChannel(conf *BaseChannelConf) *BaseChannel {
	return NewBaseChannel(conf, Default_ReadPool)
}

func NewBaseChannel(conf *BaseChannelConf, readPool *ReadPool) *BaseChannel {
	channel := &BaseChannel{
		ChannelHandle: ChannelHandle{},
		ChannelStatis: *NewChStatis(),
		chConf:        conf,
		readPool:      readPool,
		closeExit:     make(chan bool, 1),
	}
	channel.SetClosed(true)
	channel.SetChId("")
	logx.Info("create base channel, conf:", conf)
	return channel
}

func (b *BaseChannel) StartChannel(channel Channel) error {
	if !channel.IsClosed() {
		id := b.GetChId()
		return errors.New("channel is open, chId:" + id)
	}

	defer func() {
		re := recover()
		if re != nil {
			logx.Error("Start updchannel error:", re)
		}
	}()
	go b.startReadLoop(channel)

	// 启动后处理方法
	startFunc := b.HandleStartFunc
	if startFunc != nil {
		err := startFunc(channel)
		if err != nil {
			b.StopChannel(channel)
		}
		return err
	}
	logx.Info("finish to start channel, chId:", channel.GetChId())
	b.SetClosed(false)
	return nil
}

func (b *BaseChannel) NewPacket() Packet {
	panic("implement me")
}

func (b *BaseChannel) GetChId() string {
	return b.chId
}

func (b *BaseChannel) SetChId(chId string) {
	b.chId = chId
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

func (b *BaseChannel) Write(packet Packet) error {
	panic("implement me")
}

func (b *BaseChannel) GetChConf() *BaseChannelConf {
	return b.chConf
}

func (b *BaseChannel) GetChStatis() *ChannelStatis {
	return &b.ChannelStatis
}

func (b *BaseChannel) GetConn() net.Conn {
	return nil
}

func (b *BaseChannel) IsReadLoopContinued(err error) bool {
	return true
}

func (b *BaseChannel) SetHandleMsgFunc(handleMsgFunc HandleMsgFunc) {
	b.ChannelHandle.HandleMsgFunc = handleMsgFunc
}

func (b *BaseChannel) GetHandleMsgFunc() HandleMsgFunc {
	return b.ChannelHandle.HandleMsgFunc
}

func (b *BaseChannel) StopChannel(channel Channel) {
	// 关闭状态不再执行后面的内容
	if channel.IsClosed() {
		logx.Info("channel is closed, chId:", b.GetChId())
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			logx.Warn("close error, chId:", b.GetChId(), err)
		}
	}()

	logx.Info("start to close channel, chId:", b.GetChId())
	// 清理关闭相关
	b.SetClosed(true)
	b.closeExit <- true
	close(b.closeExit)
	conn := channel.GetConn()
	if conn != nil {
		conn.Close()
	}
	// 执行关闭后的方法
	// TODO 各自处理？
	closeFunc := b.HandleCloseFunc
	if closeFunc != nil {
		closeFunc(channel)
	}
	logx.Info("finish to close channel, chId:", b.GetChId())
}

// StartReadLoop 启动循环读取，读取到数据包后，放入#ReadQueue中，等待处理
func (b *BaseChannel) startReadLoop(channel Channel) {
	defer func() {
		recover()
		b.StopChannel(channel)
	}()
	logx.Info("start to readloop, chId:", b.GetChId())
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
					msgFunc := channel.GetHandleMsgFunc()
					if msgFunc != nil {
						msgFunc(rev)
					}
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

// 读取统计
func RevStatis(packet Packet) {
	statis := packet.GetChannel().GetChStatis()
	statis.RevByteNum += int64(len(packet.GetData()))
	statis.RevPacketNum += 1
	statis.RevTime = time.Now()
	logx.Debug("rev:", string(packet.GetData()))
}

// 写统计
func SendStatis(packet Packet) {
	statis := packet.GetChannel().GetChStatis()
	statis.SendByteNum += int64(len(packet.GetData()))
	statis.SendPacketNum += 1
	statis.SendTime = time.Now()
	logx.Debug("write:", string(packet.GetData()))
}
