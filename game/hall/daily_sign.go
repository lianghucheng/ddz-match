package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/db"
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
		log.Debug("签到，类型：%v，数量：%v. ", item.PropType, item.Amount)
		AddSundries(item.PropType, ud, item.Amount, db.DailySignOpt, db.DailySign, "")
	}, func() {
		user.WriteMsg(&msg.S2C_DailySign{
			Name:   item.Name,
			PropID: item.PropType,
			Amount: item.Amount,
			ImgUrl: config.GetPropBaseConfig(item.PropType).ImgUrl,
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
			if dead != 0 {
				user.GetUserData().NewDailySign = true
			}
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
	if !user.GetUserData().NewDailySign {
		cfgDs = config.GetCfgNewUserDailySignItem()
	}
	dailySignItems := []msg.DailySignItems{}
	cf := config.GetPropBaseConfig
	for i := 0; i < ud.SignTimes; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   cf((*cfgDs)[i].PropType).Name,
			PropID: (*cfgDs)[i].PropType,
			Amount: (*cfgDs)[i].Amount,
			Status: msg.SignFinish,
			ImgUrl: cf((*cfgDs)[i].PropType).ImgUrl,
		})
	}
	if !user.GetUserData().DailySign {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   cf((*cfgDs)[ud.SignTimes].PropType).Name,
			PropID: (*cfgDs)[ud.SignTimes].PropType,
			Amount: (*cfgDs)[ud.SignTimes].Amount,
			Status: msg.SignAccess,
			ImgUrl: cf((*cfgDs)[ud.SignTimes].PropType).ImgUrl,
		})
	} else {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   cf((*cfgDs)[ud.SignTimes].PropType).Name,
			PropID: (*cfgDs)[ud.SignTimes].PropType,
			Amount: (*cfgDs)[ud.SignTimes].Amount,
			Status: msg.SignDeny,
			ImgUrl: cf((*cfgDs)[ud.SignTimes].PropType).ImgUrl,
		})
	}

	for i := user.GetUserData().SignTimes + 1; i < 7; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Name:   cf((*cfgDs)[i].PropType).Name,
			PropID: (*cfgDs)[i].PropType,
			Amount: (*cfgDs)[i].Amount,
			Status: msg.SignDeny,
			ImgUrl: cf((*cfgDs)[i].PropType).ImgUrl,
		})
	}
	user.WriteMsg(&msg.S2C_DailySignItems{
		SignItems: dailySignItems,
		IsSign:    ud.DailySign,
	})
}
