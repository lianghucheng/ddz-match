package db

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
