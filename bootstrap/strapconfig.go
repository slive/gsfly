/*
 * Author:slive
 * DATE:2020/7/30
 */
package bootstrap

import (
	"gsfly/channel"
	"net/url"
)

type ServerConf interface {
	channel.AddrConf
	channel.ChannelConf
	GetMaxChannelSize() int
}

type BaseServerConf struct {
	channel.BaseAddrConf
	channel.BaseChannelConf
	MaxChannelSize int
}

func (bs *BaseServerConf) GetMaxChannelSize() int {
	return bs.MaxChannelSize
}

type ClientConf interface {
	channel.AddrConf
	channel.ChannelConf
}

type BaseClientConf struct {
	channel.BaseAddrConf
	channel.BaseChannelConf
}

type KcpConf struct {
	// TODO kcp相关的配置
}

type KcpClientConf struct {
	BaseClientConf
	KcpConf
}

type KcpServerConf struct {
	BaseServerConf
	KcpConf
}

type UdpServerConf struct {
	BaseServerConf
}

type UdpClientConf struct {
	BaseClientConf
}

type HttpxServerConf struct {
	BaseServerConf
}

type HttpxClientConf struct {
	BaseClientConf
}

type WsClientConf struct {
	BaseClientConf
	Scheme      string
	SubProtocol []string
	Path        string
	Params      map[string]interface{}
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.Scheme, Host: wsClientConf.GetAddrStr(), Path: wsClientConf.Path}
	return u.String()
}
