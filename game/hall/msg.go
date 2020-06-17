package hall

import (
	"ddz/game/player"
	"ddz/msg"
)

func UpdateUserCoupon(user *player.User) {
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
}