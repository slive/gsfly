/*
 * Author:slive
 * DATE:2020/7/30
 */
package common

import (
	"github.com/Slive/gsfly/channel"
	logx "github.com/Slive/gsfly/logger"
)

type SysConf struct {
	LogConf      *logx.LogConf
	ChannelConf  *channel.ChannelConf
	ReadPoolConf *channel.ReadPoolConf
}
