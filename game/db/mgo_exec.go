package db

import (
	"time"

	"github.com/szxby/tools/log"
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
func InsertItemLog(uid int, amount int64, item string, way string) {
	s := MongoDB.Ref()
	defer MongoDB.UnRef(s)
	data := ItemLog{
		UID:        uid,
		Item:       item,
		Way:        way,
		Amount:     amount,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err := s.DB(DB).C("itemlog").Insert(data)
	if err != nil {
		log.Error("insert fail:%v", err)
	}
}
