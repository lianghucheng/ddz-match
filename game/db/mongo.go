package db

import (
	"ddz/conf"
	"ddz/config"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/szxby/tools/log"
)

var MongoDB, BackstageDB *mongodb.DialContext

var DB,BkDBName string

func init() {
	DB = conf.GetCfgLeafSrv().DBName
	BkDBName = config.GetCfgDB().BackstageDBName
	// mongodb
	if conf.GetCfgLeafSrv().DBMaxConnNum <= 0 {
		conf.GetCfgLeafSrv().DBMaxConnNum = 100
	}
	db, err := mongodb.Dial(conf.GetCfgLeafSrv().DBUrl, conf.GetCfgLeafSrv().DBMaxConnNum)
	if err != nil {
		log.Fatal("dial mongodb error: %v", err)
	}
	MongoDB = db

	bkDB, err := mongodb.Dial(config.GetCfgDB().BkDBUrl, config.GetCfgDB().ConnNum)
	if err != nil {
		log.Fatal("the db url is: %v. dial backstage mongodb error: %v. ", config.GetCfgDB().BkDBUrl, err)
	}
	BackstageDB = bkDB
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
	err = db.EnsureCounter(DB, "counters", "flowdata")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "edyorder")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "knapsackprop")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "matchawardrecord")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "usermail")

	err = db.EnsureCounter(BkDBName, "counters", "feedback")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}

	err = db.EnsureUniqueIndex(DB, "users", []string{"accountid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
	err = db.EnsureIndex(DB, "gamerecord", []string{"userid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
	err = db.EnsureUniqueIndex(DB, "matchmanager", []string{"matchid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
	err = db.EnsureUniqueIndex(DB, "serverconfig", []string{"id"})
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

func MongoBkDBNextSeq(id string) (int, error) {
	return MongoDB.NextSeq(BkDBName, "counters", id)
}