package server

import (
	"ddz/game/db"
	"ddz/game/values"

	"github.com/szxby/tools/log"
)

func init() {
	log.Debug("init whiteList....")
	if err := db.GetWhiteList(); err != nil {
		log.Fatal("err:%v", err)
	}
	log.Debug("finish init whiteList:%+v", values.DefaultWhiteListConfig)
}

// CheckRestart 检查服务器是否重启
func CheckRestart() bool {
	return values.DefaultRestartConfig.Restartting
}

// SetRestart 获取配置
func SetRestart(restartTime int64) {
	values.DefaultRestartConfig.RestartTime = restartTime
	values.DefaultRestartConfig.Restartting = true
}

// CheckWhite 检查白名单,true代表通过检查
func CheckWhite(accountID int) bool {
	if !values.DefaultWhiteListConfig.WhiteSwitch {
		return true
	}
	for _, v := range values.DefaultWhiteListConfig.WhiteList {
		if v == accountID {
			return true
		}
	}
	return false
}
