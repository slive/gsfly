/*
 * 协程池管理类
 * Author:slive
 * DATE:2020/7/25
 */
package channel

import (
	logx "github.com/Slive/gsfly/logger"
	"hash/crc32"
	"os"
	"os/signal"
	"sync"
)

// ReadQueue 每个协程可缓冲的读队列
type ReadQueue struct {
	readChan chan IPacket
	id       int
}

// ReadPool 读协程池主要作用，用于控制读取的数据包处理个数，避免读取协程过大。
// 包含了每个协程读队列，最大支持协程数，每个协程缓冲队列可缓冲包数
type ReadPool struct {
	// 协程读队列列表
	readQueue map[int]*ReadQueue
	mut       sync.Mutex
	// 最大支持协程数
	maxReadPoolSize int
	// 每个协程缓冲队列可缓冲包数
	maxReadQueueSize int
}

// NewReadPool创建协程池
func NewReadPool(maxReadPoolSize int, maxReadQueueSize int) *ReadPool {
	r := &ReadPool{
		readQueue:        make(map[int]*ReadQueue, maxReadQueueSize),
		maxReadPoolSize:  maxReadPoolSize,
		maxReadQueueSize: maxReadQueueSize,
	}
	go func() {
		n := make(chan os.Signal, 1)
		signal.Notify(n)
		select {
		case s := <-n:
			// 结束时释放所有chan
			logx.Info("signal:", s)
			r.Close()
		}
		logx.Info("release all readchan.")
	}()
	return r
}

// Cache 放入缓冲区进行处理
func (p *ReadPool) Cache(pack IPacket) {
	id := pack.GetChannel().GetId()
	// hash方式进行分配
	key := hashCode(id) % p.maxReadPoolSize
	queue := p.fetchReadQueue(key)
	queue.readChan <- pack
}

// fetchReadQueue 获取ReadQueue，如果没有则创建
func (p *ReadPool) fetchReadQueue(key int) *ReadQueue {
	// 加锁避免获取出现问题
	p.mut.Lock()
	defer p.mut.Unlock()
	readQueue := p.get(key)
	if readQueue == nil {
		readQueue = &ReadQueue{readChan: make(chan IPacket, p.maxReadQueueSize),
			id: key}
		p.readQueue[key] = readQueue
		logx.Info("start read queue, key:", key)
		// 协程运行
		go handelReadQueue(readQueue)
	}
	return readQueue

}

// handelReadQueue 处理读队列
func handelReadQueue(queue *ReadQueue) {
	for {
		select {
		case packet, isOpened := <-queue.readChan:
			if packet != nil {
				channel := packet.GetChannel()
				context := NewChHandleContext(channel, packet)
				func() {
					handle := channel.GetChHandle()
					// 有错误可以继续执行
					defer func() {
						rec := recover()
						if rec != nil {
							logx.ErrorTracef(context, "handle message error:%v", rec)
							err, ok := rec.(error)
							if ok {
								// 捕获处理消息异常
								NotifyErrorHandle(context, err, ERR_MSG)
							}
						}
						// 释放资源
						packet.Clear()
					}()
					// 交给handle处理
					handler := handle.GetOnRead()
					handler(context)
				}()

				// 管道关闭后的操作
				if !isOpened {
					logx.InfoTracef(context, "queue handle end, queueId:%v", queue.id)
					return
				}
			}
		}
	}
}

func (p *ReadPool) get(key int) *ReadQueue {
	return p.readQueue[key]
}

// Close 关闭
func (p *ReadPool) Close() {
	p.mut.Lock()
	defer p.mut.Unlock()
	for k, queue := range p.readQueue {
		close(queue.readChan)
		delete(p.readQueue, k)
	}
}

// hashCode 哈希计算
func hashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
