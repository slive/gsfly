/*
 * Author:slive
 * DATE:2020/7/25
 */
package channel

import (
	"hash/crc32"
	"sync"
)

type ReadQueue struct {
	readChan chan Packet
}

type ReadPool struct {
	readQueue        map[int]*ReadQueue
	mut              sync.Mutex
	maxReadPoolSize  int
	maxReadQueueSize int
}

func (p *ReadPool) Cache(pack Packet) {
	id := pack.GetChannel().GetChId()
	key := hashCode(id) / p.maxReadPoolSize
	queue := p.fetchReadQueue(key)
	queue.readChan <- pack
}

func (p *ReadPool) fetchReadQueue(key int) *ReadQueue {
	p.mut.Lock()
	defer p.mut.Unlock()
	readQueue := p.get(key)
	if readQueue == nil {
		readQueue = &ReadQueue{readChan: make(chan Packet, p.maxReadQueueSize)}
		go handelReadQueue(readQueue)
	}
	p.readQueue[key] = readQueue
	return readQueue

}

func handelReadQueue(queue *ReadQueue) {
	for {
		select {
		case packet := <-queue.readChan:
			if packet != nil{
				msgFunc := packet.GetChannel().GetHandleMsgFunc()
				msgFunc(packet)
			}
		}
	}
}

func (p *ReadPool) get(key int) *ReadQueue {
	return p.readQueue[key]
}

func (p *ReadPool) Remove(id string) {
	p.mut.Lock()
	defer p.mut.Unlock()
	key := hashCode(id) / p.maxReadPoolSize
	delete(p.readQueue, key)
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

func NewReadPool(maxReadPoolSize int, maxReadQueueSize int) *ReadPool {
	r := &ReadPool{
		readQueue:        make(map[int]*ReadQueue, maxReadPoolSize),
		maxReadPoolSize:  maxReadPoolSize,
		maxReadQueueSize: maxReadQueueSize,
	}
	return r
}
