package match

import (
	"ddz/game/db"
	"encoding/json"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// 保存所有赛事列表
var (
	MatchList   = map[string]*BaseMatch{}
	UserIDMatch = map[int]*BaseMatch{}
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
	iter := s.DB(db.DB).C("match").Find(bson.M{"state": bson.M{"$eq": Signing}}).Iter()
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
			sConfig := &scoreConfig{}
			c, _ := json.Marshal(one)
			if err := json.Unmarshal(c, &sConfig); err != nil {
				log.Error("get config error:%v", err)
				return nil
			}
			NewScoreMatch(sConfig)
		default:
			log.Error("unknown match:%v", one)
		}
	}
	return nil
}
