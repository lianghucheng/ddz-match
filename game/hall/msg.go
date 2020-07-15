package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"time"
)

// UpdateUserCoupon 更新玩家点券
func UpdateUserCoupon(user *player.User, amount, before, after int64, opt int, way string) {
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
	if amount != 0 {
		game.GetSkeleton().Go(
			func() {
				data := db.ItemLog{
					UID:        user.BaseData.UserData.AccountID,
					Item:       values.Coupon,
					Way:        way,
					Amount:     amount,
					Before:     before,
					After:      after,
					OptType:    opt,
					CreateTime: time.Now().Unix(),
				}
				db.InsertItemLog(data)
			}, nil)
	}
}

func UpdateRealName(user *player.User, status int, errmsg string) {
	user.WriteMsg(&msg.S2C_RealNameAuth{
		RealName: user.RealName(),
		Error:    status,
		ErrMsg:   errmsg,
	})
}

func SendAddBankCard(user *player.User, code int, errmsg string) {
	bankCard := new(BankCard)
	bankCard.Userid = user.UID()
	bankCard.Read()
	tail := ""
	if bankCard.BankCardNo != "" {
		tail = bankCard.BankCardNo[len(bankCard.BankCardNo)-4:]
	}
	user.WriteMsg(&msg.S2C_BindBankCard{
		BankCardInfo: &msg.BankCardInfo{
			BankName:       bankCard.BankName,
			BankCardNoTail: tail,
		},
		Error:  code,
		ErrMsg: errmsg,
	})
}
