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

type Kws00ClientConf struct {
	KcpClientConf
	// Path 可选，代表所在的相对路径，用于可能存在的路由，类似http的request url，如"/admin/user"
	Path string
	// Params 可选，代表建立dial所需的参数
	Params map[string]interface{}
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
