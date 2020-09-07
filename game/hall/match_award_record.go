package hall

import (
	"ddz/config"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type MatchAwardRecord struct {
	MatchName    string
	AwardContent string
	ID           int `bson:"_id"`
	Userid       int
	Accountid    int
	MatchType    string
	MatchID      string
	CreatedAt    int64
	Realname     string
	Desc         string
	IsMoney 	 bool
}

func (ctx *MatchAwardRecord) save() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	_, err := se.DB(db.DB).C("matchawardrecord").Upsert(bson.M{"_id": ctx.ID}, ctx)
	if err != nil {
		log.Error(err.Error())
	}
}

func WriteMatchAwardRecord(uid int, matchType, matchID, matchName, awardContent string) {
	log.Debug("比赛奖励：uid: %v, matchType: %v, matchID: %v, matchName: %v, awardContent: %v. ", uid, matchType, matchID, matchType, matchName, awardContent)
	ud := player.ReadUserDataByID(uid)
	matchAwardRecord := new(MatchAwardRecord)
	matchAwardRecord.Userid = ud.UserID

	matchAwardRecord.MatchType = matchType
	matchAwardRecord.MatchID = matchID
	matchAwardRecord.CreatedAt = time.Now().Unix()
	matchAwardRecord.Realname = ud.RealName
	matchAwardRecord.Accountid = ud.AccountID
	matchAwardRecord.ID, _ = db.MongoDBNextSeq("matchawardrecord")
	matchAwardRecord.MatchName = matchName
	matchAwardRecord.AwardContent = awardContent

	if len(awardContent) >= 0 && values.GetMoneyAward(awardContent) >= 0 {
		matchAwardRecord.IsMoney = true
	}

	matchAwardRecord.save()
}

func ReadMatchAwardRecord(query bson.M) *[]MatchAwardRecord {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	datas := new([]MatchAwardRecord)


	if err := se.DB(db.DB).C("matchawardrecord").Find(query).Limit(config.GetCfgNormal().HorseLampSizeLimit).All(datas);err != nil {
		log.Error(err.Error())
	}

	return datas
}
