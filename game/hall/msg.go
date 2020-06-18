package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
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
				db.InsertItemLog(user.BaseData.UserData.UserID, amount, db.Coupon, way)
			}, nil)
	}
}
