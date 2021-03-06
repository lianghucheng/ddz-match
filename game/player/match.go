package player

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/values"
	"ddz/msg"
	"encoding/json"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// SendMatchRecordAll 先发送总的数据
func (user *User) SendMatchRecordAll() {
	uid := user.BaseData.UserData.UserID
	var items []msg.GameAllRecord
	all := []msg.OneMatchType{}
	data := db.RedisGetMatchAll(uid)
	// count := 0
	if data == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			// count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Limit(60).Count()
			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Limit(60).All(&items)
		}, func() {
			for _, r := range items {
				if c, ok := values.MatchTypeConfig[r.MatchType]; ok {
					tag := false
					for _, v := range all {
						if v.MatchType == r.MatchType {
							tag = true
							break
						}
					}
					if !tag {
						all = append(all, c)
					}
				}
			}
			user.WriteMsg(&msg.S2C_GetGameRecordAll{
				All: all,
			})
			db.RedisMatchAll(uid, items)
		})
	} else {
		err := json.Unmarshal(data, &items)
		if err != nil {
			log.Error("umarshal fail:%v", err)
			return
		}
		for _, r := range items {
			if c, ok := values.MatchTypeConfig[r.MatchType]; ok {
				tag := false
				for _, v := range all {
					if v.MatchType == r.MatchType {
						tag = true
						break
					}
				}
				if !tag {
					all = append(all, c)
				}
			}
		}
		user.WriteMsg(&msg.S2C_GetGameRecordAll{
			All: all,
		})
	}
}

// SendMatchRecord 给玩家发送战绩
func (user *User) SendMatchRecord(page, num int, matchType string) {
	uid := user.BaseData.UserData.UserID
	data := db.RedisGetMatchAll(uid)
	// 如果redis中没有数据，则去数据库中查
	var items []msg.GameAllRecord
	oneRecord := []msg.GameRecord{}
	record := &msg.S2C_GetGameRecord{}
	if data == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			// count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Limit(60).Count()
			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Limit(60).All(&items)
		}, func() {
			for _, r := range items {
				if r.MatchType != matchType {
					continue
				}
				oneRecord = append(oneRecord, msg.GameRecord{
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

			if (page-1)*num >= len(oneRecord) {
				log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(oneRecord), page, num)
				return
			}
			end := page * num
			if end > len(oneRecord) {
				end = len(oneRecord)
			}
			record.Total = len(oneRecord)
			record.PageNumber = page
			record.PageSize = num
			record.Record = oneRecord[(page-1)*num : end]
			record.MatchType = matchType
			user.WriteMsg(record)
			db.RedisMatchAll(uid, items)
		})
		return
	}
	err := json.Unmarshal(data, &items)
	if err != nil {
		log.Error("umarshal fail:%v", err)
		return
	}

	for _, r := range items {
		if r.MatchType != matchType {
			continue
		}
		oneRecord = append(oneRecord, msg.GameRecord{
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

	if (page-1)*num >= len(oneRecord) {
		log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(oneRecord), page, num)
		return
	}
	end := page * num
	if end > len(oneRecord) {
		end = len(oneRecord)
	}

	record.Total = len(oneRecord)
	record.PageNumber = page
	record.PageSize = num
	record.Record = oneRecord[(page-1)*num : end]
	record.MatchType = matchType
	user.WriteMsg(record)
}

// SendMatchRankRecord 发送某一比赛排名信息
func (user *User) SendMatchRankRecord(matchID string, page, num, rPage, rNum int) {
	uid := user.BaseData.UserData.UserID
	log.Debug("player %v get rank,page:%v,rpage:%b,rnum:%v,matchid:%v", uid, page, rPage, rNum, matchID)
	data := db.RedisGetMatchRankRecord(matchID)
	// 如果redis中没有数据，则去数据库中查
	// var items []msg.GameAllRecord
	// allRecord := &msg.SaveAllGameRecord{}
	rank := []msg.Rank{}
	if data == nil {
		match := map[string]interface{}{}
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			if err := s.DB(db.DB).C("match").Find(bson.M{"sonmatchid": matchID}).One(&match); err != nil {
				log.Error("get data err:%v", err)
				return
			}
			if match["rank"] == nil {
				log.Error("no rank in match:%v", match)
				match = nil
				return
			}
			tmp1, ok := match["rank"].([]interface{})
			if !ok {
				log.Error("no rank in match:%v", match)
				match = nil
				return
			}
			for _, v := range tmp1 {
				tmp2, ok := v.(map[string]interface{})
				if !ok {
					log.Error("no rank in match:%v", match)
					continue
				}
				jsStr, err := json.Marshal(tmp2)
				if err != nil {
					log.Error("no rank in match:%v", match)
					continue
				}
				final := msg.Rank{}
				err = json.Unmarshal(jsStr, &final)
				if err != nil {
					log.Error("no rank in match:%v", match)
					continue
				}
				rank = append(rank, final)
			}
		}, func() {
			if match == nil || len(rank) == 0 {
				return
			}
			sendData := &msg.S2C_GetGameRankRecord{
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
			db.RedisMatchRankRecord(matchID, rank)
		})
	} else {
		err := json.Unmarshal(data, &rank)
		if err != nil {
			log.Error("umarshal fail:%v", err)
			return
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

// SendMatchResultRecord 获取某一赛事的记录
func (user *User) SendMatchResultRecord(matchID string, page, num int) {
	uid := user.BaseData.UserData.UserID
	log.Debug("player %v get result,page:%v,matchid:%v", uid, page, matchID)
	redisData := db.RedisGetMatchAll(uid)
	// 如果redis中没有数据，则去数据库中查
	var items []msg.GameAllRecord
	result := []msg.GameResult{}
	if redisData == nil {
		game.GetSkeleton().Go(func() {
			s := db.MongoDB.Ref()
			defer db.MongoDB.UnRef(s)
			// count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Limit(60).Count()
			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
				Sort("-createdat").Limit(60).All(&items)
		}, func() {
			for _, r := range items {
				if r.MatchId == matchID {
					result = r.Result
					break
				}
			}
			data := &msg.S2C_GetGameResultRecord{}
			data.Total = len(result)
			data.MatchID = matchID
			data.PageNumber = page
			data.PageSize = num
			if (page-1)*num >= len(result) {
				log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), page, num)
				return
			}
			end := page * num
			if end > len(result) {
				end = len(result)
			}
			data.Result = result[(page-1)*num : end]
			user.WriteMsg(data)
			db.RedisMatchAll(uid, items)
		})
	}
	err := json.Unmarshal(redisData, &items)
	if err != nil {
		log.Error("umarshal fail:%v", err)
		return
	}
	for _, r := range items {
		if r.MatchId == matchID {
			result = r.Result
			break
		}
	}
	data := &msg.S2C_GetGameResultRecord{}
	data.Total = len(result)
	data.MatchID = matchID
	data.PageNumber = page
	data.PageSize = num
	if (page-1)*num >= len(result) {
		log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), page, num)
		return
	}
	end := page * num
	if end > len(result) {
		end = len(result)
	}
	data.Result = result[(page-1)*num : end]
	user.WriteMsg(data)
}

// SendMatchRecord 给玩家发送战绩 查询方法修改
// func (user *User) SendMatchRecord(page, num int) {
// 	uid := user.BaseData.UserData.UserID
// 	data := db.RedisGetMatchRecord(uid, page)
// 	// 如果redis中没有数据，则去数据库中查
// 	var items []msg.GameAllRecord
// 	record := &msg.S2C_GetGameRecord{}
// 	allRecord := &msg.SaveAllGameRecord{}
// 	count := 0
// 	if data == nil {
// 		game.GetSkeleton().Go(func() {
// 			s := db.MongoDB.Ref()
// 			defer db.MongoDB.UnRef(s)
// 			count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Count()
// 			if (page-1)*num >= count {
// 				log.Error("invalid page:%v,num:%v", page, num)
// 				return
// 			}
// 			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
// 				Sort("-createdat").Skip((page - 1) * num).Limit(num).All(&items)

// 		}, func() {
// 			for _, r := range items {
// 				record.Record = append(record.Record, msg.GameRecord{
// 					UserId:    r.UserId,
// 					MatchId:   r.MatchId,
// 					MatchType: r.MatchType,
// 					Desc:      r.Desc,
// 					Level:     r.Level,
// 					Award:     r.Award,
// 					Count:     r.Count,
// 					Total:     r.Total,
// 					Last:      r.Last,
// 					Wins:      r.Wins,
// 					Period:    r.Period,
// 					CreateDat: r.CreateDat,
// 				})
// 			}
// 			record.Total = count
// 			record.PageNumber = page
// 			record.PageSize = num
// 			user.WriteMsg(record)
// 			allRecord.Record = items
// 			allRecord.Total = count
// 			allRecord.PageNumber = page
// 			allRecord.PageSize = num
// 			db.RedisMatchRecord(uid, page, allRecord)
// 		})
// 	} else {
// 		err := json.Unmarshal(data, allRecord)
// 		if err != nil {
// 			log.Error("umarshal fail:%v", err)
// 			return
// 		}
// 		for _, r := range allRecord.Record {
// 			record.Record = append(record.Record, msg.GameRecord{
// 				UserId:    r.UserId,
// 				MatchId:   r.MatchId,
// 				MatchType: r.MatchType,
// 				Desc:      r.Desc,
// 				Level:     r.Level,
// 				Award:     r.Award,
// 				Count:     r.Count,
// 				Total:     r.Total,
// 				Last:      r.Last,
// 				Wins:      r.Wins,
// 				Period:    r.Period,
// 				CreateDat: r.CreateDat,
// 			})
// 		}
// 		record.Total = allRecord.Total
// 		record.PageNumber = allRecord.PageNumber
// 		record.PageSize = allRecord.PageSize
// 		user.WriteMsg(record)
// 	}
// }

// SendMatchResultRecord 发送某一比赛详细信息 查询方法修改
// func (user *User) SendMatchResultRecord(matchID string, page, num, rPage, rNum int) {
// 	uid := user.BaseData.UserData.UserID
// 	log.Debug("player %v get result,page:%v,rpage:%b,rnum:%v,matchid:%v", uid, page, rPage, rNum, matchID)
// 	data := db.RedisGetMatchRecord(uid, page)
// 	// 如果redis中没有数据，则去数据库中查
// 	var items []msg.GameAllRecord
// 	allRecord := &msg.SaveAllGameRecord{}
// 	result := []msg.GameResult{}
// 	count := 0
// 	if data == nil {
// 		game.GetSkeleton().Go(func() {
// 			s := db.MongoDB.Ref()
// 			defer db.MongoDB.UnRef(s)
// 			count, _ = s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).Count()

// 			s.DB(db.DB).C("gamerecord").Find(bson.M{"userid": user.BaseData.UserData.UserID}).
// 				Sort("-createdat").Skip((num - 1) * page).Limit(num).All(&items)

// 		}, func() {
// 			allRecord.Record = items
// 			allRecord.Total = count
// 			allRecord.PageNumber = page
// 			allRecord.PageSize = num
// 			for _, r := range allRecord.Record {
// 				if r.MatchId == matchID {
// 					result = r.Result
// 				}
// 			}
// 			sendData := msg.S2C_GetGameResultRecord{
// 				Total:      len(result),
// 				MatchID:    matchID,
// 				PageNumber: page,
// 				PageSize:   num,
// 			}
// 			if (rPage-1)*rNum >= len(result) {
// 				log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), rPage, rNum)
// 				return
// 			}
// 			end := rPage * rNum
// 			if end > len(result) {
// 				end = len(result)
// 			}
// 			sendData.Result = result[(rPage-1)*rNum : end]

// 			user.WriteMsg(sendData)
// 			// data, err := json.Marshal(allRecord)
// 			// if err != nil {
// 			// 	log.Error("marshal fail:%v", err)
// 			// 	return
// 			// }
// 			db.RedisMatchRecord(uid, page, allRecord)
// 		})
// 	} else {
// 		err := json.Unmarshal(data, allRecord)
// 		if err != nil {
// 			log.Error("umarshal fail:%v", err)
// 			return
// 		}
// 		for _, r := range allRecord.Record {
// 			if r.MatchId == matchID {
// 				result = r.Result
// 			}
// 		}
// 		data := &msg.S2C_GetGameResultRecord{}
// 		data.Total = len(result)
// 		data.MatchID = matchID
// 		data.PageNumber = rPage
// 		data.PageSize = rNum
// 		if (rPage-1)*rNum >= len(result) {
// 			log.Error("invalid params:total:%v,rpage:%v,rnum:%v", len(result), rPage, rNum)
// 			return
// 		}
// 		end := rPage * rNum
// 		if end > len(result) {
// 			end = len(result)
// 		}
// 		data.Result = result[(rPage-1)*rNum : end]

// 		user.WriteMsg(data)
// 	}
// }
