package match

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"encoding/json"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// 保存所有赛事列表
var (
	MatchList        = map[string]*BaseMatch{}
	UserIDMatch      = map[int]*BaseMatch{}
	MatchManagerList = map[string]values.MatchManager{}
	MatchConfigQueue = map[string]*values.NormalCofig{} // 后台修改赛事的队列，在比赛开始后下一场生效
)

func init() {
	if err := initMatchConfig(); err != nil {
		log.Fatal("init match fail,err:%v", err)
	}
	if err := initMatchTypeConfig(); err != nil {
		log.Fatal("init match config fail,err:%v", err)
	}
}

// match
func initMatchConfig() error {
	s := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(s)
	one := map[string]interface{}{}
	log.Debug("init MatchConfig........")
	iter := s.DB(db.DB).C("matchmanager").Find(bson.M{"state": bson.M{"$lt": Delete}}).Iter()
	for iter.Next(&one) {
		if one["matchtype"] == nil || one["matchid"] == nil {
			log.Error("unknow match:%v", one)
			continue
		}
		mType, ok := one["matchtype"].(string)
		if !ok {
			log.Error("unknow match:%v", one)
			continue
		}
		switch mType {
		case ScoreMatch, MoneyMatch, DoubleMatch, QuickMatch:
			sConfig := &ScoreConfig{}
			c, _ := json.Marshal(one)
			if err := json.Unmarshal(c, &sConfig); err != nil {
				log.Error("get config error:%v", err)
				return nil
			}
			if err := sConfig.CheckConfig(); err != nil {
				continue
			}
			// 上架时间
			if sConfig.ShelfTime > time.Now().Unix() {
				sConfig.State = Cancel
				sConfig.StartTimer = game.GetSkeleton().AfterFunc(time.Duration(sConfig.ShelfTime-time.Now().Unix())*time.Second, func() {
					// NewScoreManager(sConfig)
					sConfig.NewManager()
				})
				MatchManagerList[sConfig.MatchID] = sConfig
			} else {
				// NewScoreManager(sConfig)
				sConfig.NewManager()
			}
		default:
			log.Error("unknown match:%v", one)
		}
	}
	log.Debug("init match finish:match manager:%+v", MatchManagerList)
	return nil
}

// initMatchTypeConfig 初始化赛事类型配置
func initMatchTypeConfig() error {
	s := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(s)
	log.Debug("init MatchType Config........")
	one := msg.OneMatchType{}
	// iter := s.DB(db.DB).C("matchtypeconfig").Pipe([]bson.M{
	// 	{"$project": bson.M{
	// 		"_id": 0,
	// 	}},
	// }).Iter()
	iter := s.DB(db.DB).C("matchtypeconfig").Find(bson.M{}).Iter()
	for iter.Next(&one) {
		values.MatchTypeConfig[one.MatchType] = one
	}
	log.Debug("finish init matchtype config %v", values.MatchTypeConfig)

	return nil
}

// GetMatchManagerInfo 获取整个赛事列表的信息
func GetMatchManagerInfo(opt int) interface{} {
	matchManager := []values.MatchManager{}
	for _, v := range MatchManagerList {
		matchManager = append(matchManager, v)
	}
	sortMatch(matchManager)
	switch opt {
	case 1:
		raceInfo := []msg.RaceInfo{}
		for _, v := range matchManager {
			var award float64
			info := v.GetNormalConfig()
			// 不在大厅展示
			if !info.ShowHall || info.State != Signing {
				continue
			}
			if len(info.Award) > 0 {
				for _, v := range info.Award {
					award += values.GetMoneyAward(v)
				}
			}
			sTime := info.StartTime
			if info.StartType == 2 {
				sTime = info.ReadyTime - time.Now().Unix()
			}
			raceInfo = append(raceInfo, msg.RaceInfo{
				ID:        info.MatchID,
				Desc:      info.MatchName,
				Award:     award,
				EnterFee:  float64(info.EnterFee),
				ConDes:    info.MatchDesc,
				JoinNum:   len(info.AllSignInPlayers),
				StartTime: sTime,
				StartType: info.StartType,
				MatchType: info.MatchType,
			})
		}
		return raceInfo
	case 2:
		list := []msg.OneMatch{}
		for _, m := range matchManager {
			m := m.GetNormalConfig()
			if m.State != Signing {
				continue
			}
			sTime := m.StartTime
			if m.StartType == 2 {
				sTime = m.ReadyTime - time.Now().Unix()
			}
			list = append(list, msg.OneMatch{
				MatchID:   m.MatchID,
				MatchName: m.MatchName,
				SignInNum: len(m.AllSignInPlayers),
				Recommend: m.Recommend,
				MaxPlayer: m.MaxPlayer,
				EnterFee:  m.EnterFee,
				StartTime: sTime,
				StartType: m.StartType,
				IsSign:    false,
				MatchType: m.MatchType,
				MatchIcon: m.MatchIcon,
			})
		}
		return list
	default:
		return nil
	}
}

func sortMatch(list []values.MatchManager) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].GetNormalConfig().Sort > list[j].GetNormalConfig().Sort {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

// BroadcastMatchInfo 当赛事发生变化时，全服广播赛事信息
func BroadcastMatchInfo() {
	RaceInfo := GetMatchManagerInfo(1).([]msg.RaceInfo)
	if len(RaceInfo) == 0 {
		return
	}
	for uid, user := range player.UserIDUsers {
		if ma, ok := UserIDMatch[uid]; ok {
			myMatchID := ma.NormalCofig.MatchID
			for i, v := range RaceInfo {
				if v.ID == myMatchID {
					RaceInfo[i].IsSign = true
				} else {
					RaceInfo[i].IsSign = false
				}
			}
		}
		user.WriteMsg(&msg.S2C_RaceInfo{
			Races: RaceInfo,
		})
	}
}
