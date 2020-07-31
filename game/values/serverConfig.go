package values

import (
	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
)

// 更新状态
const (
	RestartStatusWait = iota + 1
	RestartStatusIng
	RestartStatusFinish
)

// RestartConfig 服务器重启配置
type RestartConfig struct {
	ID             string `bson:"id"`
	TipsTime       int64  `bson:"tipstime"`
	RestartTime    int64  `bson:"restarttime"`
	EndTime        int64  `bson:"endtime"`
	RestartTitle   string `bson:"restarttitle"`
	RestartType    string `bson:"restarttype"`
	Status         int    `bson:"status"`
	RestartContent string `bson:"restartcontent"`
	CreateTime     int64  `bson:"createtime"`
	RestartTimer   *timer.Timer
}

// WhiteListConfig 白名单配置
type WhiteListConfig struct {
	WhiteSwitch bool  `bson:"whiteswitch"`
	WhiteList   []int `bson:"whitelist"`
}

// 默认配置
var (
	DefaultRestartConfig   = RestartConfig{}
	DefaultWhiteListConfig = WhiteListConfig{}
)

// CheckRestart 检查服务器是否重启
func CheckRestart() bool {
	log.Debug("checkstart:%v", DefaultRestartConfig.Status)
	if DefaultRestartConfig.Status > 0 {
		return DefaultRestartConfig.Status == RestartStatusIng
	}
	return false
}
