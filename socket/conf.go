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

	// GetListenConfs 监听相关的配置
	GetListenConfs() []IListenConf
}

type IListenConf interface {
	common.IAttact

	GetNetwork() channel.Network

	// GetBasePath 监听的基本的path配置
	GetBasePath() string
}

type ListenConf struct {
	common.Attact
	network  channel.Network
	basePath string
}

func NewListenConf(network channel.Network, basePath string) *ListenConf {
	conf := ListenConf{network: network, basePath: basePath}
	conf.Attact = *common.NewAttact()
	return &conf
}

// GetBasePath 监听的基本的path配置
func (lsConf *ListenConf) GetBasePath() string {
	return lsConf.basePath
}

func (lsConf *ListenConf) GetNetwork() channel.Network {
	return lsConf.network
}

// ServerConf 服务配置
type ServerConf struct {
	channel.AddrConf
	channel.ChannelConf
	common.Id
	common.Parent
	maxChannelSize int

	listenConfs []IListenConf
}

// NewServerConf 创建服务端配置
func NewServerConf(ip string, port int, protocol channel.Network, listenConfs ...IListenConf) *ServerConf {
	s := &ServerConf{}
	s.AddrConf = *channel.NewAddrConf(ip, port)
	s.ChannelConf = *channel.NewDefChannelConf(protocol)
	s.maxChannelSize = -1
	s.listenConfs = listenConfs
	return s
}

// SetMaxChannelSize 设置可接受最大channel数
func (bs *ServerConf) SetMaxChannelSize(maxChannelSize int) {
	bs.maxChannelSize = maxChannelSize
}

// GetMaxChannelSize 获取可接受最大channel数
func (bs *ServerConf) GetMaxChannelSize() int {
	return bs.maxChannelSize
}

// GetListenConfs 监听相关的配置
func (bs *ServerConf) GetListenConfs() []IListenConf {
	return bs.listenConfs
}

// 存放ws对应的subprotocol的key值，对应的值为一个string数组
const WS_SUBPROTOCOL_KEY = "WS_SUBPROTOCOL_KEY"

type IWsServerConf interface {
	IServerConf
	GetScheme() string
}

type WsServerConf struct {
	ServerConf
	scheme string
}

func NewWsServerConf(ip string, port int, scheme string, listenConfs ...IListenConf) *WsServerConf {
	if listenConfs == nil || len(listenConfs) <= 0 {
		panic("listenConfs conf are nil.")
	}
	if len(scheme) <= 0 {
		scheme = "ws"
	}
	w := &WsServerConf{scheme: scheme}
	w.ServerConf = *NewServerConf(ip, port, channel.NETWORK_WS, listenConfs...)
	return w
}

func (wsServerConf *WsServerConf) GetScheme() string {
	return wsServerConf.scheme
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
// ip 地址
// port 端口
// reqPath 请求地址，可选
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
	GetReqPath() string
	GetSubProtocol() []string
}

type WsConf struct {
	subProtocol []string
	reqPath     string
}

func NewWsConf(reqPath string, subProtocol ...string) *WsConf {
	if len(reqPath) <= 0 {
		reqPath = ""
	}
	w := &WsConf{
		subProtocol: subProtocol,
		reqPath:     reqPath,
	}
	return w
}

func (wsConf *WsConf) GetUrl() string {
	panic("implement")
}

func (wsConf *WsConf) GetReqPath() string {
	return wsConf.reqPath
}

func (wsConf *WsConf) GetSubProtocol() []string {
	return wsConf.subProtocol
}

type IWsClientConf interface {
	IClientConf
	IWsConf
	GetUrlByPath(path string) string
	GetScheme() string
}

type WsClientConf struct {
	ClientConf
	WsConf
	scheme string
}

func NewWsClientConf(ip string, port int, scheme string, path string, subProtocol ...string) *WsClientConf {
	w := &WsClientConf{}
	w.WsConf = *NewWsConf(path, subProtocol...)
	w.ClientConf = *NewClientConf(ip, port, channel.NETWORK_WS)
	if len(scheme) <= 0 {
		scheme = "ws"
	}
	w.scheme = scheme
	return w
}

func (wsClientConf *WsClientConf) GetUrl() string {
	u := url.URL{Scheme: wsClientConf.GetScheme(), Host: wsClientConf.GetAddrStr(), Path: wsClientConf.GetReqPath()}
	return u.String()
}

func (wsClientConf *WsClientConf) GetUrlByPath(path string) string {
	u := url.URL{Scheme: wsClientConf.scheme, Host: wsClientConf.GetAddrStr(), Path: path}
	return u.String()
}

func (wsClientConf *WsClientConf) GetScheme() string {
	return wsClientConf.scheme
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
