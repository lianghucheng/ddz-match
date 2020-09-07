package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
	HorseQueue = new(HorseRaceLampQueue)
	StartHorseTimer()
}

type HorseRaceLampQueue []values.HorseRaceLamp

var (
	HorseQueue *HorseRaceLampQueue
	horseQueueIndex int
	HorseCurrTimer *timer.Timer
	HorsePrevTimer *timer.Timer
)
func SendHorseRaceLamp(username, matchName, award string) {
	log.Debug("发送跑马灯,%v", award)
	if len(award) == 0 || values.GetMoneyAward(award) == 0 {
		return
	}

	templates := config.GetCfgNormal().Templates
	if len(templates) == 0 {
		log.Error("没有配置跑马灯模版")
		return
	}

	//HorseRaceLamp := &msg.HorseRaceLamp{
	//	UserName:  username,
	//	MatchName: matchName,
	//	Amount:    values.GetMoneyAward(award),
	//}
	//info, err := utils.TransferMapStringSlice(HorseRaceLamp)
	//if err != nil {
	//	log.Error("格式转换失败")
	//	return
	//}
	player.Broadcast(&msg.S2C_HorseRaceLamp{
		Template: fmt.Sprintf(templates[rand.Intn(len(templates))], username, matchName, values.GetMoneyAward(award)),
		//Info: info,
	})
}

func (ctx *HorseRaceLampQueue)Add(lamp values.HorseRaceLamp) {
	length := len(*ctx)
	if length >= config.GetCfgNormal().HorseLampSizeLimit {
		*ctx = append([]values.HorseRaceLamp{lamp}, (*ctx)[:length - 1]...)
	} else {
		*ctx = append([]values.HorseRaceLamp{lamp}, (*ctx)...)
	}
}

func SendHorseCurrTimer(username, matchName, award string) {
	if len(award) == 0 || values.GetMoneyAward(award) == 0 {
		return
	}

	log.Debug("跑马灯定时器")

	if HorseCurrTimer != nil {
		HorseCurrTimer.Stop()
		HorseCurrTimer = nil
	}

	game.GetSkeleton().AfterFunc(8 * time.Second, func() {
		SendHorseRaceLamp(username, matchName, award)
	})

	HorseQueue.Add(values.HorseRaceLamp{
		UserName:  username,
		MatchName: matchName,
		Award:     award,
	})

	HorseCurrTimer = game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().CircleTTL) * time.Second, func() {
		horseQueueIndex = 0
		SendHorsePrevTimer()
		HorseCurrTimer = nil
	})
}

func SendHorsePrevTimer() {
	if HorseCurrTimer != nil {
		if HorsePrevTimer != nil {
			HorsePrevTimer.Stop()
			HorsePrevTimer = nil
		}
		return
	}

	if len(*HorseQueue) == 0 {
		log.Debug("*************跑马灯队列为空")
		return
	}

	if horseQueueIndex >= len(*HorseQueue) {
		horseQueueIndex = 0
	}

	username := (*HorseQueue)[horseQueueIndex].UserName
	matchName := (*HorseQueue)[horseQueueIndex].MatchName
	award := (*HorseQueue)[horseQueueIndex].Award
	SendHorseRaceLamp(username, matchName, award)
	horseQueueIndex++
	log.Debug("跑马灯缓存队列index: %v", horseQueueIndex)

	HorsePrevTimer = game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().CircleTTL) * time.Second, func() {
		SendHorsePrevTimer()
	})
}

func StartHorseTimer() {
	datas := ReadMatchAwardRecord(bson.M{"ismoney": true})
	for _, v := range *datas {
		ud := player.ReadUserDataByAid(v.Accountid)
		HorseQueue.Add(values.HorseRaceLamp{
			UserName: ud.Nickname,
			MatchName: v.MatchName,
			Award:     v.AwardContent,
		})
	}

	SendHorsePrevTimer()
}