package hall

import (
	"ddz/game"
	"ddz/game/db"
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
	ud := user.GetUserData()
	if ud.SetNickNameCount >= 1 {
		user.WriteMsg(&msg.S2C_UpdateNickName{
			Error: msg.S2C_SetNickName_More,
		})
		return
	}
	ud.Nickname = nickname
	ud.SetNickNameCount++
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
	UpdateUserCoupon(user, count, user.BaseData.UserData.Coupon-count, user.BaseData.UserData.Coupon, db.NormalOpt, db.Charge)
}

func UpdateUserAfterTaxAward(user *player.User) {
	user.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
		AfterTaxAward: utils.Decimal(user.Fee()),
	})
}

func ChangePassword(user *player.User, m *msg.C2S_ChangePassword) {
	ud := user.GetUserData()
	if ud.Password != m.OldPassword {
		user.WriteMsg(&msg.S2C_ChangePassword{
			Error: msg.ErrChangePasswordOldNo,
		})
		return
	}

	ud.Password = m.NewPassword

	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
		user.WriteMsg(&msg.S2C_ChangePassword{
			Error: msg.ErrChangePasswordSuccess,
		})
	}, nil)
}

func TakenFirstCoupon(user *player.User) {
	ud := user.GetUserData()
	ud.FirstLogin = false
	ud.Coupon += 5
	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
	}, func() {
		user.WriteMsg(&msg.S2C_TakenFirstCoupon{})
		UpdateUserCoupon(user, 5, ud.Coupon-5, ud.Coupon, db.NormalOpt, db.InitPlayer)
	})
}
