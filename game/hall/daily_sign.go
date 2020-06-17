package hall

import (
	"ddz/conf"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"time"
)

func DailySign(user *player.User) {
	checkDailySign(user)

	if user.GetUserData().DailySign {
		return
	}

	user.GetUserData().DailySign = true
	addCoupon := conf.GetCfgDailySign()[user.GetUserData().SignTimes].Chips
	user.GetUserData().Coupon += addCoupon
	user.GetUserData().SignTimes++
	player.SaveUserData(user.GetUserData())

	UpdateUserCoupon(user)
	user.WriteMsg(&msg.S2C_DailySign{
		Coupon: addCoupon,
	})

	SendDailySignItems(user)
}

func checkDailySign(user *player.User) {
	dead := user.GetUserData().DailySignDeadLine
	if dead < time.Now().Unix() {
		week := time.Unix(dead, 0).Weekday()
		dist := 0
		if week > time.Sunday {
			dist = 7 - int(week)
		}
		if week == time.Monday || time.Unix(dead, 0).Add(time.Duration(dist+1)*24*time.Hour).Unix() <= time.Now().Unix() {
			user.GetUserData().SignTimes = 0
		}

		user.GetUserData().DailySignDeadLine = utils.OneDay0ClockTimestamp(time.Now().Add(24 * time.Hour))
		user.GetUserData().DailySign = false
		player.SaveUserData(user.GetUserData())
	}
}

func SendDailySignItems(user *player.User) {
	checkDailySign(user)

	dailySignItems := []msg.DailySignItems{}
	for i := 0; i < user.GetUserData().SignTimes; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:  conf.GetCfgDailySign()[i].Chips,
			Status: msg.SignFinish,
		})
	}
	if !user.GetUserData().DailySign {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:  conf.GetCfgDailySign()[user.GetUserData().SignTimes].Chips,
			Status: msg.SignAccess,
		})
	} else {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:  conf.GetCfgDailySign()[user.GetUserData().SignTimes].Chips,
			Status: msg.SignDeny,
		})
	}

	for i := user.GetUserData().SignTimes + 1; i < 7; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:  conf.GetCfgDailySign()[i].Chips,
			Status: msg.SignDeny,
		})
	}
	user.WriteMsg(&msg.S2C_DailySignItems{
		SignItems: dailySignItems,
		IsSign:    user.GetUserData().DailySign,
	})
}
