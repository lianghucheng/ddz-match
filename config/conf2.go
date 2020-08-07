package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	//*rt = append(*rt, PriceItem{
	//	PriceID: 1,
	//	Fee:     20000,
	//	Name:    "点券",
	//	Amount:  200,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 2,
	//	Fee:     10000,
	//	Name:    "点券",
	//	Amount:  100,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 3,
	//	Fee:     5000,
	//	Name:    "点券",
	//	Amount:  50,
	//})
	*rt = append(*rt, PriceItem{
		PriceID: 4,
		Fee:     2000,
		Name:    "点券",
		Amount:  20,
	})
	// *rt = append(*rt, PriceItem{
	// 	PriceID: 5,
	// 	Fee:     1000,
	// 	Name:    "点券",
	// 	Amount:  10,
	// })
	// *rt = append(*rt, PriceItem{
	// 	PriceID: 6,
	// 	Fee:      500,
	// 	Name:    "点券",
	// 	Amount:  5,
	// })

	return rt
}

const (
	ModelDev = 1 //开发环境模式
	ModelPro = 2 //生产环境模式
)

type Config struct {
	Model                int //配置模式
	CfgMatchRobotMaxNums map[string]int
	CfgDailySignItems    *[]CfgDailySignItem
	CfgPay               *CfgPay
}

func (ctx *Config) print() {
	fmt.Printf("Model:%+v\n", ctx.Model)
	fmt.Printf("CfgMatchRobotMaxNums:%+v\n", ctx.CfgMatchRobotMaxNums)
	fmt.Printf("CfgDailySignItems:%+v\n", *ctx.CfgDailySignItems)
	fmt.Printf("CfgPay:%+v\n", *ctx.CfgPay)
}

type CfgMatchRobotMaxNum struct {
	MatchID   string //
	PerMaxNum int
	Total     int
	JoinNum   int
}

type CfgDailySignItem struct {
	ID               int
	Name             string
	IsTowardKnapsack bool
	Desc             string
	Amount           float64
}

type CfgPay struct {
	Host             string
	CreatePaymentUrl string
}

var cfg *Config

const (
	dbUrl      = "mongodb://localhost" //mongodb服务地址
	dbName     = "ddz-match"           //数据库名称
	collection = "config"              //集合名称
)

var (
	dial *mongodb.DialContext
)

func init() {
	cfg = new(Config)
	var err error
	dial, err = mongodb.Dial(dbUrl, 1)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	ret, err := ReadCfg(ModelDev)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Fatal(err.Error())
			return
		}
		b, err := ioutil.ReadFile("config/init-config.json")
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		ret = new(Config)
		if err := json.Unmarshal(b, ret); err != nil {
			log.Fatal(err.Error())
			return
		}
		ret.CfgMatchRobotMaxNums = make(map[string]int)
	}
	cfg = ret
	cfg.print()
}

func UpdateCfg(model int) error {
	ret, err := ReadCfg(model)
	if err != nil {
		log.Error("update config error: " + err.Error())
		return err
	}
	cfg = ret
	return nil
}

func UpdateCfgFromFile() error {
	return nil
}

func ReadCfg(model int) (*Config, error) {
	se := dial.Ref()
	defer dial.UnRef(se)
	cfgData := new(Config)
	if err := se.DB(dbName).C(collection).Find(bson.M{"model": model}).One(cfgData); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return cfgData, nil
}

func GetCfgDailySignItem() *[]CfgDailySignItem {
	return cfg.CfgDailySignItems
}

type TempProp struct {
	ID               int
	Name             string
	IsAdd            bool
	IsTowardKnapsack bool
	IsUse            bool
	Expiredat        int64
	Desc             string
}

const (
	PropIDAward      = 10001
	PropIDCoupon     = 10002
	PropIDCouponFrag = 10003
)

var PropList = map[int]TempProp{
	PropIDAward: {
		ID:               10001,
		Name:             "奖金",
		IsAdd:            true,
		IsTowardKnapsack: false,
		IsUse:            false,
		Expiredat:        -1,
		Desc:             "用户税后奖金超过10元可进行提现申请处理",
	},
	PropIDCoupon: {
		ID:               10002,
		Name:             "点券",
		IsAdd:            true,
		IsTowardKnapsack: false,
		IsUse:            true,
		Expiredat:        -1,
		Desc:             "用户税后奖金超过10元可进行提现申请处理",
	},
	PropIDCouponFrag: {
		ID:               10003,
		Name:             "碎片",
		IsAdd:            true,
		IsTowardKnapsack: true,
		IsUse:            true,
		Expiredat:        -1,
		Desc:             "用户税后奖金超过10元可进行提现申请处理",
	},
}

func GetCfgMatchRobotMaxNums() map[string]int {
	return cfg.CfgMatchRobotMaxNums
}

func GetCfgPay() *CfgPay {
	return cfg.CfgPay
}
