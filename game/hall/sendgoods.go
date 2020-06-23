package hall

import "ddz/game/player"

func TempPayOK(accounid, amount int) {
	ud := player.ReadUserDataByAid(accounid)
	ud.Coupon += int64(amount)
	player.SaveUserData(ud)

}