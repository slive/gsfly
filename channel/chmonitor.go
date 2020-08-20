/*
 * 监控收发信息等
 * Author:slive
 * DATE:2020/7/29
 */
package channel

import (
	"encoding/json"
	logx "github.com/Slive/gsfly/logger"
	"time"
)

type StatisUnit struct {
	// 字节数
	ByteNum   int64     `json:"byteNum"`
	Time      time.Time `json:"time"`
	SpendTime float64   `json:"spendTime"`
	IsOk      bool      `json:"isOk"`
}

func newStatisUnit() *StatisUnit {
	return &StatisUnit{Time: time.Now(), SpendTime: 0, IsOk: false, ByteNum: 0}
}

type Statis struct {
	// 总字节数
	TotalByteNum int64 `json:"totalByteNum"`
	// 总包数
	TotalPacketNum int64 `json:"totalPacketNum"`

	// 当前操作记录
	Current *StatisUnit `json:"current"`
	// 上次操作记录
	Last *StatisUnit `json:"last"`

	// 总失败字节数
	TotalFailByteNum int64 `json:"totalFailByteNum"`
	// 总失败包数
	TotalFailPacketNum int64 `json:"totalFailPacketNum"`

	// 连续失败次数
	FailTimes int64 `json:"failTimes"`
}

func newStatis() *Statis {
	return &Statis{
		TotalByteNum:       0,
		TotalPacketNum:     0,
		Current:            newStatisUnit(),
		Last:               newStatisUnit(),
		TotalFailByteNum:   0,
		TotalFailPacketNum: 0,
		FailTimes:          0,
	}
}

// ChannelStatis 统计相关，比如收发包数目，收发次数
type ChannelStatis struct {
	SendStatics      *Statis `json:"send"`
	RevStatics       *Statis `json:"rev"`
	HandleMsgStatics *Statis `json:"handleMsg"`
}

func NewChStatis() *ChannelStatis {
	return &ChannelStatis{
		SendStatics:      newStatis(),
		RevStatics:       newStatis(),
		HandleMsgStatics: newStatis(),
	}
}

func (s *ChannelStatis) StringSend() string {
	marshal, err := json.Marshal(s.SendStatics)
	if err == nil {
		return string(marshal)
	}
	return ""
}

func (s *ChannelStatis) StringRev() string {
	marshal, err := json.Marshal(s.RevStatics)
	if err == nil {
		return string(marshal)
	}
	return ""
}

func (s *ChannelStatis) StringHandle() string {
	marshal, err := json.Marshal(s.HandleMsgStatics)
	if err == nil {
		return string(marshal)
	}
	return ""
}

func copyStaticunit(current *StatisUnit, last *StatisUnit) {
	last.Time = current.Time
	last.IsOk = current.IsOk
	last.SpendTime = current.SpendTime
	last.ByteNum = current.ByteNum
}

func RevStatisFail(channel IChannel, initTime time.Time) {
	statis := channel.GetChStatis().RevStatics
	copyStaticunit(statis.Current, statis.Last)
	statis.Current.IsOk = false
	now := time.Now()
	statis.Current.Time = now
	statis.Current.SpendTime = time.Since(initTime).Seconds()
	statis.FailTimes += 1
	logx.Debug("rev fail, chId:", channel.GetId())
}

// 读取统计
func RevStatis(packet IPacket, isOk bool) {
	channel := packet.GetChannel()
	statis := channel.GetChStatis().RevStatics
	handleStatis(statis, packet, isOk)
	logx.Debugf("chId:%v, rev:%v", channel.GetId(), string(packet.GetData()))
}

func handleStatis(statis *Statis, packet IPacket, isOk bool) {
	copyStaticunit(statis.Current, statis.Last)
	dataLen := int64(len(packet.GetData()))
	statis.TotalByteNum += dataLen
	statis.TotalPacketNum += 1
	statis.Current.IsOk = isOk
	now := time.Now()
	statis.Current.Time = now
	statis.Current.SpendTime = time.Since(packet.GetInitTime()).Seconds()
	if !isOk {
		statis.TotalFailByteNum += dataLen
		statis.TotalFailPacketNum += 1
		statis.FailTimes += 1
	} else {
		statis.FailTimes = 0
		statis.Current.ByteNum += dataLen
	}
}

// 写统计
func SendStatis(packet IPacket, isOk bool) {
	channel := packet.GetChannel()
	statis := channel.GetChStatis().SendStatics
	handleStatis(statis, packet, isOk)
	logx.Debugf("chId:%v, write:%v", channel.GetId(), string(packet.GetData()))
}

// 写统计
func HandleMsgStatis(packet IPacket, isOk bool) {
	channel := packet.GetChannel()
	statis := channel.GetChStatis().HandleMsgStatics
	handleStatis(statis, packet, isOk)
	logx.Debugf("chId:%v, handle:%v", channel.GetId(), string(packet.GetData()))
}
