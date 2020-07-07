package db

import (
	"encoding/json"
	"strconv"

	"github.com/szxby/tools/log"
)

// 设定好redis的key
const (
	MatchRecordAll = "match_record_all" // 战绩总数据
	MatchRecord    = "match_record"     // 战绩
	MatchRank      = "match_rank"       // 单个比赛所有玩家的排行
)

// 数据过期时间
const expireTime = 5 * 60

// RedisMatchAll 设置赛事总数据
func RedisMatchAll(uid int, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal match record all fail:%v", err)
		return
	}
	key := strconv.Itoa(uid) + MatchRecordAll
	_, err = Do("Set", key, msg, "EX", expireTime)
	if err != nil {
		log.Error("set match record fail:%v", err)
	}
}

// RedisGetMatchAll 获取赛事总数据
func RedisGetMatchAll(uid int) []byte {
	key := strconv.Itoa(uid) + MatchRecordAll
	data, err := Do("Get", key)
	if err != nil {
		log.Error("set match record all fail:%v", err)
		return nil
	}
	ret, ok := data.([]byte)
	if !ok {
		log.Error("byte fail:%v", err)
		return nil
	}
	return ret
}

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
func RedisMatchRankRecord(matchID string, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal match record fail:%v", err)
		return
	}
	key := MatchRank + matchID
	_, err = Do("Set", key, msg, "EX", expireTime)
	if err != nil {
		log.Error("set match record fail:%v", err)
	}
}

// RedisGetMatchRankRecord 获取玩家赛事记录
func RedisGetMatchRankRecord(matchID string) []byte {
	key := MatchRank + matchID
	data, err := Do("Get", key)
	if err != nil {
		log.Error("set match record fail:%v", err)
		return nil
	}
	// log.Debug("check:%v", data)
	ret, ok := data.([]byte)
	if !ok {
		log.Error("byte fail:%v", err)
		return nil
	}
	return ret
}
