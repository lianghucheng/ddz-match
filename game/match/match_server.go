package match

import (
	"ddz/game/db"
	"ddz/game/values"
	"ddz/msg"
	"encoding/json"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// 保存所有赛事列表
var (
	MatchList        = map[string]*BaseMatch{}
	UserIDMatch      = map[int]*BaseMatch{}
	MatchManagerList = map[string]values.MatchManager{}
)

func init() {
	if err := initMatchConfig(); err != nil {
		log.Fatal("init match fail,err:%v", err)
	}
}

// match
func initMatchConfig() error {
	s := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(s)
	one := map[string]interface{}{}
	log.Debug("init MatchConfig........")
	iter := s.DB(db.DB).C("matchmanager").Find(bson.M{"state": bson.M{"$eq": Signing}}).Iter()
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
		case Score:
			sConfig := &ScoreConfig{}
			c, _ := json.Marshal(one)
			if err := json.Unmarshal(c, &sConfig); err != nil {
				log.Error("get config error:%v", err)
				return nil
			}
			NewScoreManager(sConfig)
		default:
			log.Error("unknown match:%v", one)
		}
	}
	log.Debug("init match finish:match manager:%+v", MatchManagerList)
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
			if len(info.Award) > 0 {
				award = values.ParseAward(info.Award[0])
			}
			raceInfo = append(raceInfo, msg.RaceInfo{
				ID:        info.MatchID,
				Desc:      info.MatchName,
				Award:     award,
				EnterFee:  float64(info.EnterFee) / 10,
				ConDes:    info.MatchDesc,
				JoinNum:   len(info.AllSignInPlayers),
				StartTime: info.StartTime,
			})
		}
		return raceInfo
	case 2:
		list := []msg.OneMatch{}
		for _, m := range matchManager {
			m := m.GetNormalConfig()
			list = append(list, msg.OneMatch{
				MatchID:   m.MatchID,
				MatchName: m.MatchName,
				SignInNum: len(m.AllSignInPlayers),
				Recommend: m.Recommend,
				MaxPlayer: m.MaxPlayer,
				EnterFee:  m.EnterFee,
				StartTime: m.StartTime,
				IsSign:    false,
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
