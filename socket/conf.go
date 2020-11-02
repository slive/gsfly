/*
 * 各种协议相关的配置
 * Author:slive
 * DATE:2020/7/30
 */
package socket

import (
	"github.com/Slive/gsfly/channel"
	"github.com/Slive/gsfly/common"
	"net/url"
)

// IServerConf 服务端的配置接口
type IServerConf interface {
	channel.IAddrConf
	channel.IChannelConf
	common.IId
	common.IParent
	// GetMaxChannelSize 获取可接受最大channel数
	GetMaxChannelSize() int
	// SetMaxChannelSize 设置可接受最大channel数
	SetMaxChannelSize(maxChannelSize int)
}

// ServerConf 服务配置
type ServerConf struct {
	channel.AddrConf
	channel.ChannelConf
	common.Id
	common.Parent
	MaxChannelSize int
}

// NewServerConf 创建服务端配置
func NewServerConf(ip string, port int, protocol channel.Network) *ServerConf {
	s := &ServerConf{}
	s.AddrConf = *channel.NewAddrConf(ip, port)
	s.ChannelConf = *channel.NewDefChannelConf(protocol)
	s.MaxChannelSize = 0
	return s
}

// SetMaxChannelSize 设置可接受最大channel数
func (bs *ServerConf) SetMaxChannelSize(maxChannelSize int) {
	bs.MaxChannelSize = maxChannelSize
}

// GetMaxChannelSize 获取可接受最大channel数
func (bs *ServerConf) GetMaxChannelSize() int {
	return bs.MaxChannelSize
}

// IClientConf 客户端配置接口
type IClientConf interface {
	channel.IAddrConf
	channel.IChannelConf
}

// ClientConf 客户端配置
type ClientConf struct {
	channel.AddrConf
	channel.ChannelConf
}

// NewClientConf 创建客户端配置
func NewClientConf(ip string, port int, protocol channel.Network) *ClientConf {
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
	s.ClientConf = *NewClientConf(ip, port, channel.NETWORK_KCP)
	return s
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
	s.ServerConf = *NewServerConf(ip, port, channel.NETWORK_KCP)
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
	s.ServerConf = *NewServerConf(ip, port, channel.NETWORK_UDP)
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
	s.ClientConf = *NewClientConf(ip, port, channel.NETWORK_UDP)
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

func NewWsConf(scheme string, path string, subProtocol ...string) *WsConf {
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

func NewWsServerConf(ip string, port int, scheme string, path string, subProtocol ...string) *WsServerConf {
	w := &WsServerConf{}
	w.ServerConf = *NewServerConf(ip, port, channel.NETWORK_WS)
	w.WsConf = *NewWsConf(scheme, path, subProtocol...)
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

func NewWsClientConf(ip string, port int, scheme string, path string, subProtocol ...string) *WsClientConf {
	w := &WsClientConf{}
	w.WsConf = *NewWsConf(scheme, path, subProtocol...)
	w.ClientConf = *NewClientConf(ip, port, channel.NETWORK_WS)
	return w
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.Scheme, Host: wsClientConf.GetAddrStr(), Path: wsClientConf.Path}
	return u.String()
}

type ITcpServerConf interface {
	IServerConf
}

type TcpServerConf struct {
	ServerConf
}

func NewTcpServerConf(ip string, port int) *TcpServerConf {
	s := &TcpServerConf{}
	s.ServerConf = *NewServerConf(ip, port, channel.NETWORK_TCP)
	return s
}

type ITcpClientConf interface {
	IClientConf
}

type TcpClientConf struct {
	ClientConf
}

func NewTcpClientConf(ip string, port int) *TcpClientConf {
	s := &TcpClientConf{}
	s.ClientConf = *NewClientConf(ip, port, channel.NETWORK_TCP)
	return s
}
