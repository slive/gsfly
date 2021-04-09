/*
 * Author:slive
 * DATE:2021/1/1
 */
package socket

import (
	"github.com/slive/gsfly/channel"
	logx "github.com/slive/gsfly/logger"
	"testing"
	"time"
)

func TestEchoWs(t *testing.T) {
	childConf := NewServerChildConf(channel.NETWORK_WS, "/test")
	serverConf := NewWsServerConf("127.0.0.1", 9080, "ws", childConf)
	defServerHandle := channel.NewDefChHandle(func(ctx channel.IChHandleContext) {
		logx.DebugTracef(ctx, "server msg:%v", ctx.GetPacket())
	})
	serverSocket := NewServerSocket(nil, serverConf, defServerHandle)
	serverSocket.Listen()
	if serverSocket.IsClosed() {
		logx.ErrorTracef(serverSocket, "server socket listen fail.")
		return
	}

	clientConf := NewWsClientConf("127.0.0.1", 9080, "ws", "test")
	defClientHandle := channel.NewDefChHandle(func(ctx channel.IChHandleContext) {
		logx.DebugTracef(ctx, "client msg:%v", ctx.GetPacket())
	})
	clientSocket := NewClientSocket(nil, clientConf, defClientHandle, nil)
	clientSocket.Dial()
	if clientSocket.IsClosed() {
		logx.ErrorTracef(clientSocket, "server socket dail fail.")
		return
	}

	time.Sleep(time.Second * 1)
	clientChannel := clientSocket.GetChannel()
	packet := clientChannel.NewPacket()

	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			packet.SetData([]byte("21123232"))
			clientChannel.Write(packet)
			logx.Println("21123232")
		}
	}
}

func TestEchoTcp(t *testing.T) {

}
