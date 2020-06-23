package db

import (
	"encoding/json"
	"strconv"

	"github.com/szxby/tools/log"
)

// 设定好redis的key
const (
	MatchRecord = "match_record" // 战绩
	MatchRank   = "match_rank"   // 单个比赛所有玩家的排行
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
	key := strconv.Itoa(uid) + MatchRecord + strconv.Itoa(page)
	_, err = Do("Set", key, msg, "EX", expireTime)
	if err != nil {
		log.Error("set match record fail:%v", err)
	}
}

// RedisGetMatchRecord 获取玩家赛事记录
func RedisGetMatchRecord(uid, page int) []byte {
	key := strconv.Itoa(uid) + MatchRecord + strconv.Itoa(page)
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

// RedisMatchRankRecord 设置玩家赛事记录
func RedisMatchRankRecord(uid int, matchID string, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal match record fail:%v", err)
		return
	}
	key := strconv.Itoa(uid) + MatchRank + matchID
	_, err = Do("Set", key, msg, "EX", expireTime)
	if err != nil {
		log.Error("set match record fail:%v", err)
	}
}

// RedisGetMatchRankRecord 获取玩家赛事记录
func RedisGetMatchRankRecord(uid int, matchID string) []byte {
	key := strconv.Itoa(uid) + MatchRank + matchID
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
