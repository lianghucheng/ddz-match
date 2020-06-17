package player

// SendMatchRecord 给玩家发送战绩
// func (user *User) SendMatchRecord(page, num int) {
// 	uid := user.BaseData.UserData.UserID
// 	data := db.RedisGetMatchRecord(uid, page)
// 	// 如果redis中没有数据，则去数据库中查
// 	if data == nil {

// 	} else {
// 		record := &msg.S2C_GetMatchRecord{}
// 		err := json.Unmarshal(data, record)
// 		if err != nil {
// 			log.Error("umarshal fail:%v", err)
// 			return
// 		}
// 		user.WriteMsg(record)
// 	}
// }
