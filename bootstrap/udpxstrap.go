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

func StartSimpleKwsListen(kcpServerConf *config.KcpServerConf, handleKwsFrame kcpx.HandleKwsFrame) error {
	return StartKwsListen(kcpServerConf, handleKwsFrame, nil, nil)
}

func StartKwsListen(kcpServerConf *config.KcpServerConf, handleKwsFrame kcpx.HandleKwsFrame, startFunc gch.HandleStartFunc, closeFunc gch.HandleCloseFunc) error {
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kws addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcpws error, addr:", addr, err)
		return err
	}

	kwsChannels := make(map[string]gch.Channel, 10)
	defer func() {
		for key, kch := range kwsChannels {
			kch.StopChannel()
			delete(kwsChannels, key)
		}
	}()

	for {
		conn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpwsconn error:", nil)
			return err
		}

		kwsCh := kcpx.NewKwsChannelWithHandle(conn, &kcpServerConf.ChannelConf, handleKwsFrame, startFunc, closeFunc)
		err = kwsCh.StartChannel(kwsCh)
		if err != nil {
			kwsChannels[kwsCh.GetChId()] = kwsCh
		}
	}
	return nil
}

func StartKcpListen(kcpServerConf *config.KcpServerConf, chHandle *gch.ChannelHandle) error {
	addr := kcpServerConf.GetAddrStr()
	logx.Info("listen kcp addr:", addr)
	list, err := kcp.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Info("listen kcp error, addr:", addr, err)
		return err
	}

	kwsChannels := make(map[string]gch.Channel, 10)
	defer func() {
		for key, kch := range kwsChannels {
			kch.StopChannel()
			delete(kwsChannels, key)
		}
	}()

	for {
		kcpConn, err := list.AcceptKCP()
		if err != nil {
			logx.Error("accept kcpconn error:", nil)
			return err
		}

		kcpCh := kcpx.NewKcpChannelWithHandle(kcpConn, &kcpServerConf.ChannelConf, chHandle)
		err = kcpCh.StartChannel(kcpCh)
		if err != nil {
			kwsChannels[kcpCh.GetChId()] = kcpCh
		}
	}
	return nil
}

func StartUdpListen(serverConf *config.UdpServerConf, channelHandle gch.ChannelHandle) error {
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

func DialKws(clientConf *config.KcpClientConf, handleKwsFrame kcpx.HandleKwsFrame, startFunc gch.HandleStartFunc, closeFunc gch.HandleCloseFunc) (gch.Channel, error) {
	addr := clientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return nil, err
	}
	kwsCh := kcpx.NewKwsChannelWithHandle(conn, &clientConf.ChannelConf, handleKwsFrame, startFunc, closeFunc)
	err = kwsCh.StartChannel(kwsCh)
	return kwsCh, err
}

func DialKcp(kcpClientConf *config.KcpClientConf, chHandle *gch.ChannelHandle) (gch.Channel, error) {
	addr := kcpClientConf.GetAddrStr()
	logx.Info("dial kws addr:", addr)
	conn, err := kcp.DialWithOptions(addr, nil, 0, 0)
	if err != nil {
		logx.Error("dial kcpws conn error:", nil)
		return nil, err
	}
	kcpCh := kcpx.NewKcpChannelWithHandle(conn, &kcpClientConf.ChannelConf, chHandle)
	err = kcpCh.StartChannel(kcpCh)
	return kcpCh, err
}

func DialUdp(clientConf *config.UdpClientConf, chHandle *gch.ChannelHandle) (gch.Channel, error) {
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
	udpCh := udp.NewUdpChannelWithHandle(conn, &clientConf.ChannelConf, chHandle)
	err = udpCh.StartChannel(udpCh)
	return udpCh, err
}
