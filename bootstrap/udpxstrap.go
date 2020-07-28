/*
 * Author:slive
 * DATE:2020/7/28
 */
package bootstrap

import (
	"github.com/xtaci/kcp-go"
	gch "gsfly/channel"
	"gsfly/channel/udp"
	kcpx "gsfly/channel/udp/kcp"
	"gsfly/config"
	logx "gsfly/logger"
	"net"
)

func StartKwsListen(serverConf config.KcpServerConf, handleKwsFrame kcpx.HandleKwsFrame, startFunc gch.HandleStartFunc, closeFunc gch.HandleCloseFunc) error {
	addr := serverConf.GetAddrStr()
	logx.Info("listen kws addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcpws error, addr:", addr, err)
		return err
	}

	kwsChannels := make(map[string]gch.Channel, 10)
	defer func() {
		for _, kch := range kwsChannels {
			kch.Close()
		}
	}()

	for {
		conn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpwsconn error:", nil)
			return nil
		}
		ch, _ := kcpx.StartKwsChannelWithHandle(conn, &serverConf.ChannelConf, handleKwsFrame, startFunc, closeFunc)
		kwsChannels[ch.GetChId()] = ch
	}
	return nil
}
func StartKcpListen(serverConf config.KcpServerConf, channelHandle gch.ChannelHandle) error {
	addr := serverConf.GetAddrStr()
	logx.Info("listen kcp addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcp error, addr:", addr, err)
		return err
	}

	kwsChannels := make(map[string]gch.Channel, 10)
	defer func() {
		for _, kch := range kwsChannels {
			kch.Close()
		}
	}()

	for {
		conn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpconn error:", nil)
			return nil
		}
		ch, _ := kcpx.StartKcpChannelWithHandle(conn, &serverConf.ChannelConf, channelHandle)
		kwsChannels[ch.GetChId()] = ch
	}
	return nil
}

func StartUdpListen(serverConf config.UdpServerConf, channelHandle gch.ChannelHandle) error {
	addr := serverConf.GetAddrStr()
	logx.Info("dial udp addr:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.Error("resolve updaddr error:", err)
		return err
	}
	list, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logx.Error("listen upd error:", err)
		return err
	}
	// TODO udp有源和目标地址之分，待实现
	list.Close()
	// for {
	// conn, err := list.
	// if err != nil {
	// 	logx.Error("accept kcpconn error:", nil)
	// 	return nil
	// }
	// ch, _ := udp.StartUdpChannelWithHandle(conn, &serverConf.ChannelConf, channelHandle)
	// }
	return nil

}

func DialKws(clientConf config.KcpClientConf, handleKwsFrame kcpx.HandleKwsFrame, startFunc gch.HandleStartFunc, closeFunc gch.HandleCloseFunc) (gch.Channel, error) {
	addr := clientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return nil, err
	}
	return kcpx.StartKwsChannelWithHandle(conn, &clientConf.ChannelConf, handleKwsFrame, startFunc, closeFunc)
}

func DialKcp(clientConf config.KcpClientConf, channelHandle gch.ChannelHandle) (gch.Channel, error) {
	addr := clientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return nil, err
	}
	return kcpx.StartKcpChannelWithHandle(conn, &clientConf.ChannelConf, channelHandle)
}

func DialUdp(clientConf config.UdpClientConf, channelHandle gch.ChannelHandle) (gch.Channel, error) {
	addr := clientConf.GetAddrStr()
	logx.Info("dial udp addr:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logx.Error("resolve updaddr error:", err)
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logx.Error("dial udp conn error:", nil)
		return nil, err
	}
	return udp.StartUdpChannelWithHandle(conn, &clientConf.ChannelConf, channelHandle)
}
