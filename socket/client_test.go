/*
 * Author:slive
 * DATE:2021/1/1
 */
package socket

import (
	"github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
	"testing"
	"time"
)

func TestWsClient(t *testing.T) {

	clientConf := NewWsClientConf("127.0.0.1", 9080, "ws", "/test")
	defClientHandle := channel.NewDefChHandle(func(ctx channel.IChHandleContext) {
		logx.DebugTracef(ctx, "client msg:%v", ctx.GetPacket())
	})

	clientSocket := NewClientSocket(nil, clientConf, defClientHandle, nil)
	clientSocket.Dial()
	if clientSocket.IsClosed() {
		return
	}

	time.Sleep(time.Second * 1)
	clientChannel := clientSocket.GetChannel()
	packet := clientChannel.NewPacket()

	packet.SetData([]byte("21123232"))
	clientChannel.Write(packet)
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			logx.Println("11111")
		}
	}
}
