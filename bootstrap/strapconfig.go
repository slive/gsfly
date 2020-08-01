/*
 * Author:slive
 * DATE:2020/7/30
 */
package bootstrap

import (
	"gsfly/channel"
	"net/url"
)

type ServerConf struct {
	channel.AddrConf
	channel.ChannelConf
	MaxAcceptSize int
}

type ClientConf struct {
	channel.AddrConf
	channel.ChannelConf
}


type KcpConf struct {
	// TODO kcp相关的配置
}


type KcpClientConf struct {
	ClientConf
	KcpConf
}

type KcpServerConf struct {
	ServerConf
	KcpConf
}

type UdpServerConf struct {
	ServerConf
}

type UdpClientConf struct {
	ClientConf
}

type HttpxServerConf struct {
	ServerConf
}

type HttpxClientConf struct {
	ClientConf
}

type WsClientConf struct {
	ClientConf
	Scheme      string
	SubProtocol []string
	Path        string
	Params      map[string]interface{}
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.Scheme, Host: wsClientConf.GetAddrStr(), Path: wsClientConf.Path}
	return u.String()
}
