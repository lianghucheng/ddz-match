package server

import (
	"ddz/game/db"
	"ddz/game/match"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"

	"github.com/szxby/tools/log"
)

func init() {
	log.Debug("init whiteList....")
	if err := db.GetWhiteList(); err != nil {
		log.Fatal("err:%v", err)
	}
	log.Debug("finish init whiteList:%+v", values.DefaultWhiteListConfig)
	if err := db.GetRestart(); err != nil {
		log.Error("err:%v", err)
	}
	log.Debug("finish init restart:%+v", values.DefaultRestartConfig)
}

// KickAllPlayers 服务器更新,踢出所有玩家
func KickAllPlayers() {
	// 报名中的玩家先全部踢出
	for _, m := range match.MatchManagerList {
		m.CloseMatch()
	}
	for _, u := range player.UserIDUsers {
		// 不在比赛中踢出
		if _, ok := match.UserIDMatch[u.BaseData.UserData.UserID]; !ok {
			u.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_ServerRestart, Info: values.DefaultRestartConfig})
			u.Close()
			delete(player.UserIDUsers, u.BaseData.UserData.UserID)
		}
	}
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
