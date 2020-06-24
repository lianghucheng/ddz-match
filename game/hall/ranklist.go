package hall

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"fmt"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	RankGameTypeAward = 1
	RankGameTypeChips = 2
)

type Rank struct {
	UserID   int
	GameType int
	Nickname string
	JoinNum  int
	WinNum   int
	FailNum  int
	Award    float64
}

func FlushRank(gameType int, rankType string, userid int, value float64) {
	cfghall := conf.GetCfgHall()

	rank := new(Rank)
	rank.UserID = userid
	rank.GameType = gameType
	rank.readByUserID()
	switch rankType {
	case cfghall.RankTypeJoinNum:
		rank.JoinNum++
	case cfghall.RankTypeWinNum:
		rank.WinNum++
	case cfghall.RankTypeFailNum:
		rank.FailNum++
	case cfghall.RankTypeAward:
		rank.Award += value
	}
	rank.save()
}

func (ctx *Rank) save() {
	game.GetSkeleton().Go(func() {
		se := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(se)
		_, err := se.DB(db.DB).C("rank").Upsert(bson.M{"userid": ctx.UserID}, ctx)
		if err != nil {
			log.Error(err.Error())
		}
	}, nil)
}

func (ctx *Rank) readByUserID() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	err := se.DB(db.DB).C("rank").Find(bson.M{"userid": ctx.UserID, "gametype": ctx.GameType}).One(ctx)
	if err != nil {
		log.Error(err.Error())
	}
}

func (ctx *Rank) read(rankType string) *[]Rank {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	rt := new([]Rank)
	var err error

	cfghall := conf.GetCfgHall()
	switch rankType {
	case cfghall.RankTypeJoinNum:
		err = se.DB(db.DB).C("rank").Find(bson.M{"gametype": ctx.GameType}).Sort("-joinnum").Limit(20).All(rt)
	case cfghall.RankTypeWinNum:
		err = se.DB(db.DB).C("rank").Find(bson.M{"gametype": ctx.GameType}).Sort("-winnum").Limit(20).All(rt)
	case cfghall.RankTypeAward:
		err = se.DB(db.DB).C("rank").Find(bson.M{"gametype": ctx.GameType}).Sort("-award").Limit(20).All(rt)
	case cfghall.RankTypeFailNum:
		err = se.DB(db.DB).C("rank").Find(bson.M{"gametype": ctx.GameType}).Sort("-failnum").Limit(20).All(rt)
	}
	if err != nil {
		log.Error(err.Error())
	}
	return rt
}

func SendRankingList(user *player.User) {
	user.WriteMsg(&msg.S2C_RankingList{
		AwardRankingList: *rankByTypes(user, RankGameTypeAward),
	})
}

func rankByTypes(user *player.User, gametype int) *[]msg.RankByType {
	cfghall := conf.GetCfgHall()
	rank := new(Rank)
	rank.GameType = gametype
	mine := new(Rank)
	mine.GameType = gametype
	mine.UserID = user.BaseData.UserData.UserID
	rankByTypes := []msg.RankByType{}
	for _, title := range cfghall.RankingTitle {
		rt := rank.read(title)
		mine.readByUserID()
		rankByTypes = append(rankByTypes, msg.RankByType{
			Name:  title,
			Ranks: *rankings(rt, title),
			Mine: msg.Ranking{
				Order:    mineOrder(mine, rt),
				NickName: user.BaseData.UserData.Nickname,
				Value:    stringValue(title, mine),
			},
		})
	}
	return &rankByTypes
}

func rankings(rt *[]Rank, title string) *[]msg.Ranking {
	rankings := []msg.Ranking{}
	for k, v := range *rt {
		userData := player.ReadUserDataByID(v.UserID)
		rankings = append(rankings, msg.Ranking{
			Order:    k + 1,
			NickName: userData.Nickname,
			Value:    stringValue(title, &v),
		})
	}
	return &rankings
}

func stringValue(title string, rank *Rank) string {
	cfghall := conf.GetCfgHall()
	value := ""
	switch title {
	case cfghall.RankTypeJoinNum:
		value = fmt.Sprintf("%v", rank.JoinNum)
	case cfghall.RankTypeWinNum:
		value = fmt.Sprintf("%v", rank.WinNum)
	case cfghall.RankTypeAward:
		value = fmt.Sprintf("%v", rank.Award)
	case cfghall.RankTypeFailNum:
		value = fmt.Sprintf("%v", rank.FailNum)
	}
	return value
}

func mineOrder(mine *Rank, allRank *[]Rank) int {
	for k, v := range *allRank {
		if mine.UserID == v.UserID {
			return k + 1
		}
	}
	return -1
}
