package db

import (
	"encoding/json"
	"strconv"

	"github.com/szxby/tools/log"
)

// 设定好redis的key
const (
	MatchRecord = "match_record" // 战绩
)

// 数据过期时间
const expireTime = 5 * 60

// RedisMatchRecord 设置玩家赛事记录
func RedisMatchRecord(uid, page int, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal match record fail:%v", err)
		return
	}
	key := MatchRecord + strconv.Itoa(uid) + strconv.Itoa(page)
	_, err = Do("Set", key, msg)
	if err != nil {
		log.Error("set match record fail:%v", err)
	}
}

// RedisGetMatchRecord 获取玩家赛事记录
func RedisGetMatchRecord(uid, page int) []byte {
	key := MatchRecord + strconv.Itoa(uid) + strconv.Itoa(page)
	data, err := Do("Get", key)
	if err != nil {
		log.Error("set match record fail:%v", err)
		return nil
	}
	ret, ok := data.([]byte)
	if !ok {
		log.Error("byte fail:%v", err)
		return nil
	}
	return ret
}
