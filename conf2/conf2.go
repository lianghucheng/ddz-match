package conf2

import (
	"github.com/name5566/leaf/db/mongodb"
	"github.com/name5566/leaf/log"
)

type Config2 struct {
	RmbCouponRate int
	NewGiftCoupon int
}

type PriceItem struct {
	PriceID int
	Fee     int64
	Name    string
	Amount  int
}

var dbcfg Config2

//func DBCfgInit() {
//	se := db.MongoDB.Ref()
//	defer db.MongoDB.UnRef(se)
//	if err := se.DB(db.DB).C("conf").Find(nil).One(&dbcfg); err != nil {
//		log.Fatal(err.Error())
//	}
//	log.Debug("【数据库配置文件】%v", dbcfg)
//}

func GetCouponRate() int {
	return dbcfg.RmbCouponRate
}

func GetGiftCoupon() int {
	return dbcfg.RmbCouponRate
}

func GetPriceMenu() *[]PriceItem {
	rt := new([]PriceItem)
	*rt = append(*rt, PriceItem{
		PriceID: 1,
		Fee:     20000,
		Name:    "点券",
		Amount:  200,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 2,
		Fee:     10000,
		Name:    "点券",
		Amount:  100,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 3,
		Fee:     5000,
		Name:    "点券",
		Amount:  50,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 4,
		Fee:     2000,
		Name:    "点券",
		Amount:  20,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 5,
		Fee:     1000,
		Name:    "点券",
		Amount:  10,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 6,
		Fee:     500,
		Name:    "点券",
		Amount:  5,
	})

	return rt
}

const (
	ModelDev = iota //开发环境模式
	ModelPro //生产环境模式
)

type CfgMatchRobotMaxNum struct {
	MatchID string //
	MaxNum int
}

type Config struct {
	Model int //配置模式

}

const (
	dbUrl = "mongodb://localhost" //mongodb服务地址
	dbName = "ddz-match" //数据库名称
	collection = "conf" //集合名称
)

var (
	dial *mongodb.DialContext
)

func init() {
	var err error
	dial, err = mongodb.Dial(dbUrl, 1)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	se := dial.Ref()
	defer dial.UnRef(se)
	se.DB(dbName).C(collection).Find(nil)
}

func initCfg() error {
	return nil
}

func ReadCfg() {
	se := dial.Ref()
	defer dial.UnRef(se)

	se.DB(dbName).C(collection).Find(nil).One(nil)
}