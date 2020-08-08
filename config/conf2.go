package config

import (
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/db/mongodb"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
)

type Config2 struct {
	RmbCouponRate int
	NewGiftCoupon int
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

const (
	ModelDev = 1 //开发环境模式
	ModelPro = 2 //生产环境模式
)

type Config struct {
	Model                int //配置模式
	CfgMatchRobotMaxNums map[string]int
	CfgDailySignItems    *[]CfgDailySignItem
	CfgPay map[string]*CfgPay
	CfgDB *CfgDB
}

func (ctx *Config)print() {
	buf ,err := json.Marshal(ctx)
	if err != nil {
		log.Error(err.Error())
		return
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(buf, &m); err != nil {
		log.Error(err.Error())
		return
	}
	for k, v := range m {
		fmt.Println(k, v)
	}
}

type CfgMatchRobotMaxNum struct {
	MatchID string //
	PerMaxNum  int
	Total   int
	JoinNum int
}

type CfgDailySignItem struct {
	ID               int
	Name             string
	IsTowardKnapsack bool
	Desc             string
	Amount           float64
}

type CfgPay struct {
	NotifyHost string
	NotifyUrl string
	PayHost string
	CreatePaymentUrl string
	AppID     int
	AppToken  string
	AppSecret string
}

type CfgDB struct {
	GameDBName string
	BackstageDBName string
	BkDBUrl string
	ConnNum int
}

var cfg *Config

const (
	dbUrl      = "mongodb://192.168.1.8" //mongodb服务地址
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

func GetCfgPay() map[string]*CfgPay {
	return cfg.CfgPay
}

func GetCfgDB() *CfgDB {
	return cfg.CfgDB
}