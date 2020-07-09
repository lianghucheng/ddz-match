package conf2

import (
	"ddz/game/db"
	"github.com/name5566/leaf/log"
)

type Config2 struct {
	RmbCouponRate int
	NewGiftCoupon int
}

var dbcfg Config2

func DBCfgInit() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	if err := se.DB(db.DB).C("conf").Find(nil).One(&dbcfg); err != nil {
		log.Fatal(err.Error())
	}
	log.Debug("【数据库配置文件】%v", dbcfg)
}

func GetCouponRate() int {
	return dbcfg.RmbCouponRate
}

func GetGiftCoupon() int {
	return dbcfg.RmbCouponRate
}
