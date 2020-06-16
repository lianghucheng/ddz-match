package hall

import (
	"ddz/conf"
	"ddz/game/player"
	"ddz/msg"
)

func SendFirstRecharge(user *player.User) {
	user.WriteMsg(&msg.S2C_FirstRechage{
		Gifts: conf.GetCfgFirstRechage(),
		Money: conf.GetCfgLeafSrv().FirstRecharge,
	})
}
