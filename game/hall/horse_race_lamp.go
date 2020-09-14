package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
	HorseQueue = new(HorseRaceLampQueue)
	LevelHorseLampMap[1] = make(IDHorseLampMap)
	LevelHorseLampMap[2] = make(IDHorseLampMap)
	LevelHorseLampMap[3] = make(IDHorseLampMap)
	LevelHorseLampMap[4] = make(IDHorseLampMap)
	InitHorseTimer()
}

type HorseRaceLampQueue []values.HorseRaceLamp

var (
	HorseQueue      *HorseRaceLampQueue
	horseQueueIndex int
	//HorseCurrTimer  *timer.Timer
	//HorsePrevTimer  *timer.Timer
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

func (ctx *HorseRaceLampQueue) Add(lamp values.HorseRaceLamp) {
	length := len(*ctx)
	if length >= config.GetCfgNormal().HorseLampSizeLimit {
		*ctx = append([]values.HorseRaceLamp{lamp}, (*ctx)[:length-1]...)
	} else {
		*ctx = append([]values.HorseRaceLamp{lamp}, (*ctx)...)
	}
}

func SendHorseCurrTimer(username, matchName, award string) {
	if len(award) == 0 || values.GetMoneyAward(award) == 0 {
		return
	}

	log.Debug("跑马灯定时器---暂时的加队列")

	//if HorseCurrTimer != nil {
	//	HorseCurrTimer.Stop()
	//	HorseCurrTimer = nil
	//}
	//
	//game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().EndRoundHorseTTL)*time.Second, func() {
	//	SendHorseRaceLamp(username, matchName, award)
	//})

	HorseQueue.Add(values.HorseRaceLamp{
		UserName:  username,
		MatchName: matchName,
		Award:     award,
	})

	//HorseCurrTimer = game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().CircleTTL)*time.Second, func() {
	//	log.Debug("开启轮播老数据定时器")
	//	horseQueueIndex = 0
	//	HorseCurrTimer = nil
	//	SendHorsePrevTimer()
	//})
}

func SendHorsePrevTimer() {
	//if HorseCurrTimer != nil {
	//	if HorsePrevTimer != nil {
	//		HorsePrevTimer.Stop()
	//		HorsePrevTimer = nil
	//	}
	//	return
	//}
	//
	//if len(*HorseQueue) == 0 {
	//	log.Debug("*************跑马灯队列为空")
	//	return
	//}
	//
	//if horseQueueIndex >= len(*HorseQueue) {
	//	horseQueueIndex = 0
	//}
	//
	//username := (*HorseQueue)[horseQueueIndex].UserName
	//matchName := (*HorseQueue)[horseQueueIndex].MatchName
	//award := (*HorseQueue)[horseQueueIndex].Award
	//SendHorseRaceLamp(username, matchName, award)
	//horseQueueIndex++
	//log.Debug("跑马灯缓存队列index: %v", horseQueueIndex)
	//
	//HorsePrevTimer = game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().CircleTTL)*time.Second, func() {
	//	SendHorsePrevTimer()
	//})
}

func InitHorseTimer() {
	datas := ReadMatchAwardRecord(bson.M{"ismoney": true})
	for _, v := range *datas {
		ud := player.ReadUserDataByAid(v.Accountid)
		HorseQueue.Add(values.HorseRaceLamp{
			UserName:  ud.Nickname,
			MatchName: v.MatchName,
			Award:     v.AwardContent,
		})
	}

	//SendHorsePrevTimer()
	StartHorseTimer()
}

var IDHorseTimer = make(map[int]*timer.Timer)
var IDTakeEffectHorseTimer = make(map[int]*timer.Timer)

type IDHorseLampMap map[int]*msg.RPC_HorseLamp

var LevelHorseLampMap = make(map[int]IDHorseLampMap)
var IDLevelMap = make(map[int]int)

func StartHorseLamp(lamp *msg.RPC_HorseLamp) {
	log.Debug("StartHorseLamp")
	CleanHorseCommand(lamp.ID)
	IDLevelMap[lamp.ID] = lamp.Level
	horseLamp, ok := LevelHorseLampMap[lamp.Level]
	if !ok {
		log.Debug("no LevelHorseLampMap")
		LevelHorseLampMap[lamp.Level] = make(IDHorseLampMap)
		horseLamp = LevelHorseLampMap[lamp.Level]
	}
	horseLamp[lamp.ID] = lamp
	//now := time.Now().Unix()
	//if int(now) >= lamp.TakeEffectAt {
	//	new_horseTimer := CircleHorseTimer(lamp)
	//	IDHorseTimer[lamp.ID] = new_horseTimer
	//} else {
	//	IDTakeEffectHorseTimer[lamp.ID] = game.GetSkeleton().AfterFunc(time.Duration(lamp.TakeEffectAt - int(now)), func() {
	//		new_horseTimer := CircleHorseTimer(lamp)
	//		IDHorseTimer[lamp.ID] = new_horseTimer
	//	})
	//}

	db.SaveBkHorseLamp(lamp.ID, 0)
	log.Debug("StartHorseLamp开始成功")
}

func CircleHorseTimer(lamp *msg.RPC_HorseLamp) *timer.Timer {
	return game.GetSkeleton().AfterFunc(time.Duration(lamp.Duration)*time.Second, func() {
		if int(time.Now().Unix()) > lamp.ExpiredAt {
			CleanHorseCommand(lamp.ID)
			return
		}

		//log.Debug("HorseControl BroadCast S--->C: %v", msg.S2C_HorseRaceLamp{
		//	Template:    lamp.Template,
		//	LinkMatchID: lamp.LinkMatchID,
		//})
		player.Broadcast(&msg.S2C_HorseRaceLamp{
			Template:    lamp.Template,
			LinkMatchID: lamp.LinkMatchID,
		})
		CircleHorseTimer(lamp)
	})
}

func CleanHorseCommand(lampID int) {
	log.Debug("********CleanHorseCommand")
	//horse_timer, ok := IDHorseTimer[timerid]
	//if ok {
	//	horse_timer.Stop()
	//	delete(IDHorseTimer, timerid)
	//}
	//takeEffectTimer, ok := IDTakeEffectHorseTimer[timerid]
	//if ok {
	//	takeEffectTimer.Stop()
	//	delete(IDTakeEffectHorseTimer, timerid)
	//}
	level, ok := IDLevelMap[lampID]
	if !ok {
		log.Debug("no IDLevelMap")
		return
	}

	delete(IDLevelMap, lampID)

	idHorseLamp, ok := LevelHorseLampMap[level]
	if !ok {
		log.Debug("no LevelHorseLampMap")
		return
	}
	delete(idHorseLamp, lampID)

	db.SaveBkHorseLamp(lampID, 1)
	log.Debug("清理成功")
}

func StopHorseLamp(timerid int) {
	log.Debug("StopHorseLamp")
	CleanHorseCommand(timerid)
	log.Debug("StopHorseLamp暂停成功")
}

func StartHorseTimer() {
	log.Debug("跑马灯优先级定时器")

	game.GetSkeleton().AfterFunc(time.Duration(config.GetCfgNormal().CircleTTL)*time.Second, func() {
		template, linkMatchID := GetHorseLampDataByLevel()
		if template == "" {
			SendHorsePrev()
		} else {
			//log.Debug("HorseControl BroadCast S--->C: %v", msg.S2C_HorseRaceLamp{
			//	Template:    template,
			//	LinkMatchID: linkMatchID,
			//})
			player.Broadcast(&msg.S2C_HorseRaceLamp{
				Template:    template,
				LinkMatchID: linkMatchID,
			})
		}

		//PrintHorseLampMapNextTmp()
		StartHorseTimer()
	})
}

func GetHorseLampDataByLevel() (string, string) {
	log.Debug("开始获取优先级跑马灯")
	var horseLampData *msg.RPC_HorseLamp
	flag := 0
	for i := 1; ; {
		log.Debug("循环获取优先级跑马灯  index %v   LevelHorseLampMap长度：%v", i, len(LevelHorseLampMap))
		horseLampMap, ok := LevelHorseLampMap[i]
		if !ok {
			log.Debug("没有优先级跑马灯   index:%v", i)
			break
		}
		ids := GetMapSortKey(horseLampMap)
		log.Debug("得到排序   sort：%v", ids)
		now := int(time.Now().Unix())
		for _, v := range ids {
			data := horseLampMap[v]
			if data.TakeEffectAt < now && now < data.ExpiredAt {
				if data.NextTmp > now {
					continue
				}

				horseLampData = data
				flag = 1
				data.NextTmp = now + data.Duration
				log.Debug("优先级跑马灯开始生效！！！")
				break
			} else if data.ExpiredAt <= now {
				log.Debug("优先级跑马灯已过期")
				CleanHorseCommand(data.ID)
			}
		}
		i++
		if flag == 1 {
			break
		}
	}

	if horseLampData == nil {
		return "", ""
	}
	log.Debug("最终获取的跑马灯数据:  %+v", *horseLampData)
	return horseLampData.Template, horseLampData.LinkMatchID
}

func GetMapSortKey(m IDHorseLampMap) []int {
	ids := []int{}
	for k := range m {
		log.Debug("Map里的key %v", k)
		ids = append(ids, k)
	}
	sort.Ints(ids)
	return ids
}

func SendHorsePrev() {
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
}

func PrintHorseLampMapNextTmp() {
	for i := 1; ; {
		log.Debug("循环获取优先级跑马灯  index %v   LevelHorseLampMap长度：%v", i, len(LevelHorseLampMap))
		horseLampMap, ok := LevelHorseLampMap[i]
		if !ok {
			log.Debug("没有优先级跑马灯   index:%v", i)
			break
		}
		ids := GetMapSortKey(horseLampMap)
		log.Debug("得到排序   sort：%v", ids)
		now := int(time.Now().Unix())
		for _, v := range ids {
			data := horseLampMap[v]
			if data.TakeEffectAt < now && now < data.ExpiredAt {
				log.Debug("优先级为%v， 内容为%v，   NextTmp：%v", i, data.Template, time.Unix(int64(data.NextTmp), 0).Format("2006/01/02 15:04:05"))
			} else if data.ExpiredAt <= now {
				log.Debug("优先级跑马灯已过期")
			}
		}
		i++
	}
}
