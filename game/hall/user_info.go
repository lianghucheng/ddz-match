package hall

import (
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"gopkg.in/mgo.v2/bson"
)

func SetNickname(user *player.User, nickname string) {
	if len(nickname) < 3 || len(nickname) > 18 {
		user.WriteMsg(&msg.S2C_UpdateNickName{
			Error: msg.S2C_SetNickName_Length,
		})
		return
	}
	user.BaseData.UserData.Nickname = nickname
	player.UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"nickname": nickname}})
	user.WriteMsg(&msg.S2C_UpdateNickName{
		Error:    0,
		NickName: nickname,
	})
}

func AddCoupon(user *player.User, count int64) {
	user.BaseData.UserData.Coupon += count
	user.WriteMsg(&msg.S2C_GetCoupon{
		Error: msg.S2C_GetCouponSuccess,
	})
	UpdateUserCoupon(user)
}

func UpdateUserAfterTaxAward(user *player.User, value float64) {
	user.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
		AfterTaxAward: utils.Decimal(value),
	})
}