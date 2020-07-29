/*
 * 协程池管理类
 * Author:slive
 * DATE:2020/7/25
 */
package channel

import (
	logx "gsfly/logger"
	"hash/crc32"
	"runtime"
	"sync"
)

// ReadQueue 每个协程可缓冲的读队列
type ReadQueue struct {
	readChan chan Packet
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
	return r
}

// NewDefaultReadPool 创建默认的协程池
// 默认maxReadPoolSize = runtime.NumCPU() * 100
// 默认maxReadQueueSize int = 100
func NewDefaultReadPool() *ReadPool {
	// 与可用CPU关联
	var maxReadPoolSize int = runtime.NumCPU() * 100
	var maxReadQueueSize int = 100
	return NewReadPool(maxReadPoolSize, maxReadQueueSize)
}

// Cache 放入缓冲区进行处理
func (p *ReadPool) Cache(pack Packet) {
	id := pack.GetChannel().GetChId()
	// hash方式进行分配
	key := hashCode(id) % p.maxReadPoolSize
	queue := p.fetchReadQueue(key)
	queue.readChan <- pack
}

// fetchReadQueue 获取ReadQueue， 如果没有则创建
func (p *ReadPool) fetchReadQueue(key int) *ReadQueue {
	// 加锁避免获取出现问题
	p.mut.Lock()
	defer p.mut.Unlock()
	readQueue := p.get(key)
	if readQueue == nil {
		readQueue = &ReadQueue{readChan: make(chan Packet, p.maxReadQueueSize)}
		p.readQueue[key] = readQueue
		logx.Info("start read queue, key:", key)
		// 协程运行
		go handelReadQueue(readQueue)
	}
	return readQueue

}

func handelReadQueue(queue *ReadQueue) {
	for {
		select {
		case packet := <-queue.readChan:
			if packet != nil {
				msgFunc := packet.GetChannel().GetHandleMsgFunc()
				if msgFunc != nil {
					msgFunc(packet)
				}
			}
			break
		}
	}
}

func (p *ReadPool) get(key int) *ReadQueue {
	return p.readQueue[key]
}

func (p *ReadPool) Close() {
	p.mut.Lock()
	defer p.mut.Unlock()
	for k, queue := range p.readQueue {
		close(queue.readChan)
		delete(p.readQueue, k)
	}
}

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
