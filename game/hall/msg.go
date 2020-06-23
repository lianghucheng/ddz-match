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
		RealName: user.RealName(),
		Error:    status,
	})
}

func SendAddBankCard(user *player.User, code int) {
	bankCard := new(BankCard)
	bankCard.Userid = user.UID()
	bankCard.Read()
	tail := ""
	if  bankCard.BankCardNo != "" {
		tail = bankCard.BankCardNo[len(bankCard.BankCardNo)-4:]
	}
	user.WriteMsg(&msg.S2C_BindBankCard{
		BankCardInfo:&msg.BankCardInfo{
			BankName:bankCard.BankName,
			BankCardNoTail:tail,
		},
		Error:      code,
	})
}
