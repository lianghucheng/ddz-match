package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"github.com/szxby/tools/log"
	"time"
)

func DailySign(user *player.User) {
	checkDailySign(user)
	ud := user.GetUserData()
	if user.GetUserData().DailySign {
		return
	}

	ud.DailySign = true
	cfgDs := config.GetCfgDailySignItem()
	item := (*cfgDs)[ud.SignTimes]
	ud.SignTimes++
	game.GetSkeleton().Go(func() {
		log.Debug("签到，类型：%v，数量：%v. ", item.ID, item.Amount)
		switch item.ID {
		case config.PropTypeAward:
			WriteFlowData(ud.UserID, item.Amount, FlowTypeSign, "", "", []int{})
			ud.Fee = FeeAmount(ud.UserID)
			player.SaveUserData(ud)
			UpdateUserAfterTaxAward(user)
		case config.PropTypeCoupon:
			ud.Coupon += int64(item.Amount)
			player.SaveUserData(user.GetUserData())
		case config.PropTypeCouponFrag:
			AddPropAmount(item.ID, ud.AccountID, int(item.Amount))
		}
	}, func() {
		user.WriteMsg(&msg.S2C_DailySign{
			Name:   item.Name,
			PropID: item.ID,
			Amount: item.Amount,
		})

		SendDailySignItems(user)
	})
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
	ud := user.GetUserData()
	cfgDs := config.GetCfgDailySignItem()
	dailySignItems := []msg.DailySignItems{}
	for i := 0; i < ud.SignTimes; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   (*cfgDs)[i].Name,
			PropID: (*cfgDs)[i].ID,
			Amount: (*cfgDs)[i].Amount,
			Status: msg.SignFinish,
		})
	}
	if !user.GetUserData().DailySign {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   (*cfgDs)[ud.SignTimes].Name,
			PropID: (*cfgDs)[ud.SignTimes].ID,
			Amount: (*cfgDs)[ud.SignTimes].Amount,
			Status: msg.SignAccess,
		})
	} else {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   (*cfgDs)[ud.SignTimes].Name,
			PropID: (*cfgDs)[ud.SignTimes].ID,
			Amount: (*cfgDs)[ud.SignTimes].Amount,
			Status: msg.SignDeny,
		})
	}

	for i := user.GetUserData().SignTimes + 1; i < 7; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   (*cfgDs)[i].Name,
			PropID: (*cfgDs)[i].ID,
			Amount: (*cfgDs)[i].Amount,
			Status: msg.SignDeny,
		})
	}
	user.WriteMsg(&msg.S2C_DailySignItems{
		SignItems: dailySignItems,
		IsSign:    ud.DailySign,
	})
}
