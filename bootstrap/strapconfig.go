/*
 * Author:slive
 * DATE:2020/7/30
 */
package bootstrap

import (
	"gsfly/channel"
	"net/url"
)

type IServerConf interface {
	channel.IAddrConf
	channel.IChannelConf
	GetMaxChannelSize() int
}

type ServerConf struct {
	channel.AddrConf
	channel.ChannelConf
	MaxChannelSize int
}

func (bs *ServerConf) GetMaxChannelSize() int {
	return bs.MaxChannelSize
}

type IClientConf interface {
	channel.IAddrConf
	channel.IChannelConf
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

type Kws00ClientConf struct {
	KcpClientConf
	// Path 可选，代表所在的相对路径，用于可能存在的路由，类似http的request url，如"/admin/user"
	Path string
	// Params 可选，代表建立dial所需的参数
	Params map[string]interface{}
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
