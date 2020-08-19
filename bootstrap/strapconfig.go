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

func NewServerConf(ip string, port int, protocol channel.Protocol) *ServerConf {
	s := &ServerConf{}
	s.AddrConf = *channel.NewAddrConf(ip, port)
	s.ChannelConf = *channel.NewDefChannelConf(protocol)
	s.MaxChannelSize = 0
	return s
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

func NewClientConf(ip string, port int, protocol channel.Protocol) *ClientConf {
	s := &ClientConf{}
	s.AddrConf = *channel.NewAddrConf(ip, port)
	s.ChannelConf = *channel.NewDefChannelConf(protocol)
	return s
}

type IKcpConf interface {
	// TODO kcp相关的配置
}

type KcpConf struct {
	// TODO kcp相关的配置
}

type IKcpClientConf interface {
	IClientConf
	IKcpConf
}

type KcpClientConf struct {
	ClientConf
	KcpConf
}

func NewKcpClientConf(ip string, port int) *KcpClientConf {
	s := &KcpClientConf{}
	s.ClientConf = *NewClientConf(ip, port, channel.PROTOCOL_KCP)
	return s
}

type IKws00ClientConf interface {
	IKcpClientConf
	GetPath() string
}

type Kws00ClientConf struct {
	KcpClientConf
	// Path 可选，代表所在的相对路径，用于可能存在的路由，类似http的request url，如"/admin/user"
	Path string
}

func NewKws00ClientConf(ip string, port int, path string) *Kws00ClientConf {
	s := &Kws00ClientConf{}
	s.ClientConf = *NewClientConf(ip, port, channel.PROTOCOL_KWS00)
	s.Path = path
	return s
}

func (kwsClientConf *Kws00ClientConf) GetPath() string {
	return kwsClientConf.Path
}

type IKcpServerConf interface {
	IServerConf
	IKcpConf
}

type KcpServerConf struct {
	ServerConf
	KcpConf
}

func NewKcpServerConf(ip string, port int) *KcpServerConf {
	s := &KcpServerConf{}
	s.ServerConf = *NewServerConf(ip, port, channel.PROTOCOL_KCP)
	return s
}

type IKw00ServerConf interface {
	IKcpServerConf
}

type Kw00ServerConf struct {
	KcpServerConf
}

func NewKw00ServerConf(ip string, port int) *Kw00ServerConf {
	s := &Kw00ServerConf{}
	s.ServerConf = *NewServerConf(ip, port, channel.PROTOCOL_KWS00)
	return s
}

type IUdpServerConf interface {
	IServerConf
}

type UdpServerConf struct {
	ServerConf
}

func NewUdpServerConf(ip string, port int) *UdpServerConf {
	s := &UdpServerConf{}
	s.ServerConf = *NewServerConf(ip, port, channel.PROTOCOL_UDP)
	return s
}

type IUdpClientConf interface {
	IClientConf
}

type UdpClientConf struct {
	ClientConf
}

func NewUdpClientConf(ip string, port int) *UdpClientConf {
	s := &UdpClientConf{}
	s.ClientConf = *NewClientConf(ip, port, channel.PROTOCOL_UDP)
	return s
}

type IWsConf interface {
	GetUrl() string
	GetPath() string
	GetSubProtocol() []string
	GetScheme() string
}

type WsConf struct {
	Scheme      string
	SubProtocol []string
	Path        string
}

func NewWsConf(scheme string, path string, subProtocol []string) *WsConf {
	w := &WsConf{
		Scheme:      scheme,
		SubProtocol: subProtocol,
		Path:        path,
	}
	return w
}

func (wsConf *WsConf) GetUrl() string {
	panic("implement")
}

func (wsConf *WsConf) GetPath() string {
	return wsConf.Path
}

func (wsConf *WsConf) GetSubProtocol() []string {
	return wsConf.SubProtocol
}

func (wsConf *WsConf) GetScheme() string {
	return wsConf.Scheme
}

type IWsServerConf interface {
	IServerConf
	IWsConf
}

type WsServerConf struct {
	ServerConf
	WsConf
}

func NewWsServerConf(ip string, port int, scheme string, path string, subProtocol []string) *WsServerConf {
	w := &WsServerConf{}
	w.ServerConf = *NewServerConf(ip, port, channel.PROTOCOL_WS)
	w.WsConf = *NewWsConf(scheme, path, subProtocol)
	return w
}

func (wsServerConf *WsServerConf) GetUrl() string {
	u := url.URL{Scheme: wsServerConf.Scheme, Host: wsServerConf.GetAddrStr(), Path: wsServerConf.Path}
	return u.String()
}

type IWsClientConf interface {
	IClientConf
	IWsConf
}

type WsClientConf struct {
	ClientConf
	WsConf
}

func NewWsClientConf(ip string, port int, scheme string, path string, subProtocol []string) *WsClientConf {
	w := &WsClientConf{}
	w.WsConf = *NewWsConf(scheme, path, subProtocol)
	w.ClientConf = *NewClientConf(ip, port, channel.PROTOCOL_WS)
	return w
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.Scheme, Host: wsClientConf.GetAddrStr(), Path: wsClientConf.Path}
	return u.String()
}
