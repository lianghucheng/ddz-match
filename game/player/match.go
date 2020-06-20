package player

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/msg"
	"encoding/json"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// SendMatchRecord 给玩家发送战绩
func (user *User) SendMatchRecord(page, num int) {
	uid := user.BaseData.UserData.UserID
	data := db.RedisGetMatchRecord(uid, page)
	// 如果redis中没有数据，则去数据库中查
	var items []msg.GameAllRecord
	record := &msg.S2C_GetGameRecord{}
	allRecord := &msg.SaveAllGameRecord{}
	count := 0
	if data == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Count()
			if (page-1)*num >= count {
				log.Error("invalid page:%v,num:%v", page, num)
				return
			}
			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Skip((page - 1) * num).Limit(num).All(&items)

		}, func() {
			for _, r := range items {
				record.Record = append(record.Record, msg.GameRecord{
					UserId:    r.UserId,
					MatchId:   r.MatchId,
					MatchType: r.MatchType,
					Desc:      r.Desc,
					Level:     r.Level,
					Award:     r.Award,
					Count:     r.Count,
					Total:     r.Total,
					Last:      r.Last,
					Wins:      r.Wins,
					Period:    r.Period,
					CreateDat: r.CreateDat,
				})
			}
			record.Total = count
			record.PageNumber = page
			record.PageSize = num
			user.WriteMsg(record)
			allRecord.Record = items
			allRecord.Total = count
			allRecord.PageNumber = page
			allRecord.PageSize = num
			db.RedisMatchRecord(uid, page, allRecord)
		})
	} else {
		err := json.Unmarshal(data, allRecord)
		if err != nil {
			log.Error("umarshal fail:%v", err)
			return
		}
		for _, r := range allRecord.Record {
			record.Record = append(record.Record, msg.GameRecord{
				UserId:    r.UserId,
				MatchId:   r.MatchId,
				MatchType: r.MatchType,
				Desc:      r.Desc,
				Level:     r.Level,
				Award:     r.Award,
				Count:     r.Count,
				Total:     r.Total,
				Last:      r.Last,
				Wins:      r.Wins,
				Period:    r.Period,
				CreateDat: r.CreateDat,
			})
		}
		record.Total = allRecord.Total
		record.PageNumber = allRecord.PageNumber
		record.PageSize = allRecord.PageSize
		user.WriteMsg(record)
	}
}

// SendMatchRankRecord 发送某一比赛排名信息
func (user *User) SendMatchRankRecord(matchID string, page, num, rPage, rNum int) {
	uid := user.BaseData.UserData.UserID
	log.Debug("player %v get rank,page:%v,rpage:%b,rnum:%v,matchid:%v", uid, page, rPage, rNum, matchID)
	data := db.RedisGetMatchRecord(uid, page)
	// 如果redis中没有数据，则去数据库中查
	var items []msg.GameAllRecord
	allRecord := &msg.SaveAllGameRecord{}
	rank := []msg.Rank{}
	count := 0
	if data == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Count()

			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Skip((num - 1) * page).Limit(num).All(&items)

		}, func() {
			allRecord.Record = items
			allRecord.Total = count
			allRecord.PageNumber = page
			allRecord.PageSize = num
			for _, r := range allRecord.Record {
				if r.MatchId == matchID {
					rank = r.Rank
					break
				}
			}
			sendData := msg.S2C_GetGameRankRecord{
				Total:      len(rank),
				MatchID:    matchID,
				PageNumber: page,
				PageSize:   num,
			}
			if (rPage-1)*rNum >= len(rank) {
				log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(rank), rPage, rNum)
				return
			}
			end := rPage * rNum
			if end > len(rank) {
				end = len(rank)
			}
			sendData.Rank = rank[(rPage-1)*rNum : end]

			user.WriteMsg(sendData)
			data, err := json.Marshal(allRecord)
			if err != nil {
				log.Error("marshal fail:%v", err)
				return
			}
			db.RedisMatchRecord(uid, page, data)
		})
	} else {
		err := json.Unmarshal(data, allRecord)
		if err != nil {
			log.Error("umarshal fail:%v", err)
			return
		}
		for _, r := range allRecord.Record {
			if r.MatchId == matchID {
				rank = r.Rank
				break
			}
		}
		data := &msg.S2C_GetGameRankRecord{}
		data.Total = len(rank)
		data.PageNumber = rPage
		data.PageSize = rNum
		data.MatchID = matchID
		if (rPage-1)*rNum >= len(rank) {
			log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(rank), rPage, rNum)
			return
		}
		end := rPage * rNum
		if end > len(rank) {
			end = len(rank)
		}
		data.Rank = rank[(rPage-1)*rNum : end]

		user.WriteMsg(data)
	}
}

// SendMatchResultRecord 发送某一比赛详细信息
func (user *User) SendMatchResultRecord(matchID string, page, num, rPage, rNum int) {
	uid := user.BaseData.UserData.UserID
	log.Debug("player %v get result,page:%v,rpage:%b,rnum:%v,matchid:%v", uid, page, rPage, rNum, matchID)
	data := db.RedisGetMatchRecord(uid, page)
	// 如果redis中没有数据，则去数据库中查
	var items []msg.GameAllRecord
	allRecord := &msg.SaveAllGameRecord{}
	result := []msg.GameResult{}
	count := 0
	if data == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Count()

			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Skip((num - 1) * page).Limit(num).All(&items)

		}, func() {
			allRecord.Record = items
			allRecord.Total = count
			allRecord.PageNumber = page
			allRecord.PageSize = num
			for _, r := range allRecord.Record {
				if r.MatchId == matchID {
					result = r.Result
				}
			}
			sendData := msg.S2C_GetGameResultRecord{
				Total:      len(result),
				MatchID:    matchID,
				PageNumber: page,
				PageSize:   num,
			}
			if (rPage-1)*rNum >= len(result) {
				log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), rPage, rNum)
				return
			}
			end := rPage * rNum
			if end > len(result) {
				end = len(result)
			}
			sendData.Result = result[(rPage-1)*rNum : end]

			user.WriteMsg(sendData)
			data, err := json.Marshal(allRecord)
			if err != nil {
				log.Error("marshal fail:%v", err)
				return
			}
			db.RedisMatchRecord(uid, page, data)
		})
	} else {
		err := json.Unmarshal(data, allRecord)
		if err != nil {
			log.Error("umarshal fail:%v", err)
			return
		}
		for _, r := range allRecord.Record {
			if r.MatchId == matchID {
				result = r.Result
			}
		}
		data := &msg.S2C_GetGameResultRecord{}
		data.Total = len(result)
		data.MatchID = matchID
		data.PageNumber = rPage
		data.PageSize = rNum
		if (rPage-1)*rNum >= len(result) {
			log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), rPage, rNum)
			return
		}
		end := rPage * rNum
		if end > len(result) {
			end = len(result)
		}
		data.Result = result[(rPage-1)*rNum : end]

		user.WriteMsg(data)
	}
}
