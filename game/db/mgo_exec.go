package db

import (
	"ddz/game/values"

	"gopkg.in/mgo.v2"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// MgoGetMatchRecord 获取玩家战绩
// func MgoGetMatchRecord(uid, page, num int) {
// 	s := MongoDB.Ref()
// 	defer MongoDB.UnRef(s)
// 	iter := s.DB(DB).C("gamerecord").Find(bson.M{"userid": uid}).Iter()
// 	data := &msg.S2C_GetMatchRecord{}
// 	one := values.DDZGameRecord{}
// 	for iter.Next(&one) {
// 		data.RecordList = append(data.RecordList, one)
// 	}
// }

// InsertItemLog 插入变动日志
func InsertItemLog(data ItemLog) {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	err := s.DB(DB).C("itemlog").Insert(data)
	if err != nil {
		log.Error("insert fail:%v", err)
	}
}

// InsertMatchRecord 插入玩家单次比赛战绩
func InsertMatchRecord(record values.DDZGameRecord) {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	err := db.DB(DB).C("gamerecord").Insert(record)
	if err != nil {
		log.Error("save gamerecord %v data error: %v", record, err)
	}
}

// UpdateMatchManager 修改比赛赛事配置数据
func UpdateMatchManager(matchID string, update interface{}) error {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	_, err := db.DB(DB).C("matchmanager").Upsert(bson.M{"matchid": matchID}, update)
	if err != nil {
		log.Error("update match manager %v update: %v error: %v", matchID, update, err)
		return err
	}
	return nil
}

// GetUserMatchReview 获取玩家赛事总览数据
func GetUserMatchReview(uid int, matchType, matchID string) (values.UserMatchReview, error) {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	one := values.UserMatchReview{}
	err := db.DB(DB).C("matchreview").Find(bson.M{"uid": uid, "matchtype": matchType, "matchid": matchID}).One(&one)
	if err != nil && err != mgo.ErrNotFound {
		log.Error("err:%v", err)
		return one, err
	}
	return one, nil
}

// UpsertUserMatchReview 更新玩家赛事总览数据
func UpsertUserMatchReview(selector interface{}, update interface{}) error {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	_, err := db.DB(DB).C("matchreview").Upsert(selector, update)
	if err != nil {
		log.Error("err:%v", err)
		return err
	}
	return nil
}

// UpdateBankInfo 更新银行卡信息
func UpdateBankInfo(uid int, update interface{}) error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	if _, err := s.DB(DB).C("bankcard").Upsert(bson.M{"userid": uid}, update); err != nil && err != mgo.ErrNotFound {
		log.Error("err:%v", err)
		return err
	}
	return nil
}

func Save(coll string, data interface{}, selector bson.M) {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	_, err := se.DB(DB).C(coll).Upsert(selector, data)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func Read(coll string, data interface{}, query bson.M) {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if err := se.DB(DB).C(coll).Find(query).One(data); err != nil {
		log.Error(err.Error())
		return
	}
}