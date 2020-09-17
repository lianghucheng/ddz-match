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

// InsertIllegalMatchRecord 插入玩家异常赛事记录
func InsertIllegalMatchRecord(record values.IllegalGameRecord) {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	err := db.DB(DB).C("gameillegalrecord").Insert(record)
	if err != nil {
		log.Error("save gamerecord %v data error: %v", record, err)
	}
}

// UpdateIllegalMatchRecord 更新异常赛事
func UpdateIllegalMatchRecord(selector interface{}, update interface{}) error {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)

	err := db.DB(DB).C("gameillegalrecord").Update(selector, update)
	if err != nil {
		return err
	}
	return nil
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

func ReadAll(coll string, data interface{}, query bson.M) {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if err := se.DB(DB).C(coll).Find(query).All(data); err != nil {
		log.Error(err.Error())
		return
	}
}

// GetUserGameData 获取玩家游戏数据
func GetUserGameData(uid int) *values.GameData {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	data := &values.GameData{}
	if err := s.DB(DB).C("gamedata").Find(bson.M{"uid": uid}).One(data); err != nil && err != mgo.ErrNotFound {
		log.Error("err:%v", err)
		return nil
	}
	return data
}

// UpsertUserGameData 更新玩家游戏数据
func UpsertUserGameData(selector interface{}, update interface{}) error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	if _, err := s.DB(DB).C("gamedata").Upsert(selector, update); err != nil {
		log.Error("err:%v", err)
		return err
	}
	return nil
}

// GetWhiteList 获取白名单
func GetWhiteList() error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	wConfig := values.WhiteListConfig{}
	if err := s.DB(DB).C("serverconfig").Find(bson.M{"config": "whitelist"}).One(&wConfig); err != nil {
		log.Error("err:%v", err)
		return err
	}
	values.DefaultWhiteListConfig = wConfig
	return nil
}

// UpdateWhite 更新白名单状态
func UpdateWhite(open bool) error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	if err := s.DB(DB).C("serverconfig").Update(bson.M{"config": "whitelist"}, bson.M{"$set": bson.M{"whiteswitch": open}}); err != nil {
		log.Error("err:%v", err)
		return err
	}
	return nil
}

// GetRestart 获取重启配置
func GetRestart() error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	rConfig := values.RestartConfig{}
	if err := s.DB(DB).C("serverconfig").Find(bson.M{"config": "restart"}).
		Sort("-createtime").Limit(1).One(&rConfig); err != nil && err != mgo.ErrNotFound {
		log.Error("err:%v", err)
		return err
	}
	values.DefaultRestartConfig = rConfig
	return nil
}

// UpdateRestart 更新重启配置
func UpdateRestart(selector interface{}, update interface{}) error {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	if err := s.DB(DB).C("serverconfig").Update(selector, update); err != nil {
		log.Error("err:%v", err)
		return err
	}
	return nil
}

func ReadFlowDataByID(id int) *values.FlowData {
	query := bson.M{"_id": id}
	flowData := new(values.FlowData)
	readOneByQuery(flowData, query, "flowdata")
	return flowData
}

func readOneByQuery(rt interface{}, query bson.M, coll string) {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if err := se.DB(DB).C(coll).Find(query).One(rt); err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
	}
}

func ReadFlowdataLateOver(latedAt int64, accountid int) *values.FlowData {
	data := new(values.FlowData)
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if err := se.DB(DB).C("flowdata").Find(bson.M{"accountid": accountid, "createdat": bson.M{"$gt": latedAt}, "flowtype": 2, "status": 2}).One(data); err != nil {
		log.Debug(err.Error())
		return nil
	}
	return data
}

func ReadFlowdataBack(start, end int64, accountid int) *[]values.FlowData {
	datas := new([]values.FlowData)
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if err := se.DB(DB).C("flowdata").Find(bson.M{"accountid": accountid, "createdat": bson.M{"$gt": start, "$lt": end}, "flowtype": 2, "status": 3}).All(datas); err != nil {
		log.Debug(err.Error())
		return nil
	}
	return datas
}

func SaveFlowdata(data *values.FlowData) {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	if _, err := se.DB(DB).C("flowdata").Upsert(bson.M{"_id": data.ID}, data); err != nil {
		log.Debug(err.Error())
	}
}

// InsertLoginLog 插入登录日志
func InsertLoginLog(loginLog values.LoginLog) {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	if err := s.DB(DB).C("loginlog").Insert(loginLog); err != nil {
		log.Error("err:%v", err)
	}
}
