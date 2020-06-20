package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
)

// UpdateUserCoupon 更新玩家点券
func UpdateUserCoupon(user *player.User, amount int64, way string) {
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
	if amount != 0 {
		game.GetSkeleton().Go(
			func() {
				db.InsertItemLog(user.BaseData.UserData.UserID, amount, values.Coupon, way)
			}, nil)
	}
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
