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

func UpdateRealName(user *player.User, status int) {
	user.WriteMsg(&msg.S2C_RealNameAuth{
		RealName:user.RealName(),
		Error: status,
	})
}

func SendBankCard(user *player.User) {
	user.WriteMsg(&msg.S2C_BankCard{
		BankCardNo:user.BankCardNo(),
	})
}

func SendAddBankCard(user *player.User, code int) {
	user.WriteMsg(&msg.S2C_AddBankCard{
		BankCardNo:user.BankCardNo(),
		Error:code,
	})
}