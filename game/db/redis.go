package db

import (
	"ddz/conf"
	"time"

	"github.com/szxby/tools/log"

	"github.com/garyburd/redigo/redis"
)

func init() {
	log.Debug("连接redis服务器")
	CacheInit(conf.GetCfgRedis().Address, conf.GetCfgRedis().Password, conf.GetCfgRedis().Db)
}

var pool *redis.Pool

// 连接登陆redis 服务器
func CacheInit(server, password string, db int) {
	pool = &redis.Pool{
		MaxIdle:     500,
		IdleTimeout: 600 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				server,
				redis.DialDatabase(db),
				redis.DialPassword(password),
				redis.DialConnectTimeout(0),
				redis.DialReadTimeout(time.Second),
				redis.DialWriteTimeout(time.Second))
			if err != nil {
				log.Fatal("redis error")
				return nil, err
			}
			log.Debug("处理redis服务器:%v", err)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	pool.Get()
}

func Send(cmd string, args ...interface{}) error {
	red := pool.Get()
	defer red.Close()
	err := red.Send(cmd, args...)
	if err != nil {
		return err
	}
	return red.Flush()
}

func Do(cmd string, args ...interface{}) (interface{}, error) {
	red := pool.Get()
	defer red.Close()
	return red.Do(cmd, args...)
}
