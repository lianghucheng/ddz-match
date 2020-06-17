package db

import (
	"ddz/conf"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/szxby/tools/log"
)

var MongoDB *mongodb.DialContext

var DB string

func init() {
	DB = conf.GetCfgLeafSrv().DBName
	// mongodb
	if conf.GetCfgLeafSrv().DBMaxConnNum <= 0 {
		conf.GetCfgLeafSrv().DBMaxConnNum = 100
	}
	db, err := mongodb.Dial(conf.GetCfgLeafSrv().DBUrl, conf.GetCfgLeafSrv().DBMaxConnNum)
	if err != nil {
		log.Fatal("dial mongodb error: %v", err)
	}
	MongoDB = db
	initCollection()
}

func initCollection() {
	db := MongoDB
	err := db.EnsureCounter(DB, "counters", "users")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "configs")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "totalresult")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "roundresult")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "mailbox")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "usermail")
	err = db.EnsureUniqueIndex(DB, "users", []string{"accountid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
	err = db.EnsureIndex(DB, "gamerecord", []string{"userid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
}

func MongoDBDestroy() {
	MongoDB.Close()
	MongoDB = nil
}

func MongoDBNextSeq(id string) (int, error) {
	return MongoDB.NextSeq(DB, "counters", id)
}
