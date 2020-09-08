package config

import (
	"ddz/game/values"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	Model                    int //配置模式
	CfgMatchRobotMaxNums     map[string]int
	CfgDailySignItems        *[]CfgDailySignItem
	CfgPay                   map[string]*CfgPay
	CfgDB                    *CfgDB
	CfgPropBases             map[int]*CfgPropBase
	CfgLianHang              *CfgLianHang
	CfgNewUserDailySignItems *[]CfgDailySignItem
	CfgNormal                *CfgNormal
}

func (ctx *Config) print() {
	log.Debug("config *******************配置信息")
	buf, err := json.Marshal(ctx)
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
	MatchID   string //
	PerMaxNum int
	Total     int
	JoinNum   int
}

type CfgDailySignItem struct {
	PropType         int
	Name             string
	IsTowardKnapsack bool
	Desc             string
	Amount           float64
}

type CfgPay struct {
	NotifyHost       string
	NotifyUrl        string
	PayHost          string
	CreatePaymentUrl string
	AppID            int
	AppToken         string
	AppSecret        string
	PayType          int
}

type CfgDB struct {
	GameDBName      string
	BackstageDBName string
	BkDBUrl         string
	ConnNum         int
}

type CfgPropBase struct {
	PropType int    //道具类型, 1是点券，2是奖金，3点券碎片
	Name     string //名称
	ImgUrl   string //图片url
}

//天眼数聚
type CfgLianHang struct {
	AppKey      string
	AppSecret   string
	AppCode     string
	Host        string
	LianHangUrl string
}

type CfgNormal struct {
	AmountLimit float64
	Templates   []string
	CircleTTL   int
	HorseLampSizeLimit int
	EndRoundHorseTTL int
}

var cfg *Config
var BaseCfg *BaseConfig

type BaseConfig struct {
	DBUrl      string //mongodb服务地址
	DBName     string //数据库名称
	Collection string //集合名称
}

var (
	dial *mongodb.DialContext
)

func init() {
	var (
		err     error
		baseCfg *BaseConfig
		baseBuf []byte
	)
	baseCfg = new(BaseConfig)
	baseBuf, err = ioutil.ReadFile("config/base-config.json")
	if err != nil {
		log.Fatal("read base config fail. error: ", err.Error())
	}
	err = json.Unmarshal(baseBuf, baseCfg)
	if err != nil {
		log.Fatal("parse struct fail. error: ", err.Error())
	}
	BaseCfg = baseCfg
	cfg = new(Config)
	dial, err = mongodb.Dial(baseCfg.DBUrl, 1)
	if err != nil {
		log.Fatal("Read config from mongodb fail. err: ", err.Error())
		return
	}
	ret, err := ReadCfg(ModelDev)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Fatal("Read config from mongodb, but there was an unexpected error. the error is: ", err.Error())
			return
		}
		b, err := ioutil.ReadFile("config/init-config.json")
		if err != nil {
			log.Fatal("Read config from config.json, the error is: error: ", err.Error())
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

func UpdateCfg2() error {
	b, err := ioutil.ReadFile("config/init-config.json")
	if err != nil {
		log.Fatal("Read config from config.json, the error is: error: ", err.Error())
		return err
	}
	ret := new(Config)
	if err := json.Unmarshal(b, ret); err != nil {
		log.Fatal(err.Error())
		return err
	}
	ret.CfgMatchRobotMaxNums = make(map[string]int)
	cfg = ret
	cfg.print()
	return nil
}

func UpdateCfgFromFile() error {
	return nil
}

func ReadCfg(model int) (*Config, error) {
	se := dial.Ref()
	defer dial.UnRef(se)
	cfgData := new(Config)
	if err := se.DB(BaseCfg.DBName).C(BaseCfg.Collection).Find(bson.M{"model": model}).One(cfgData); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return cfgData, nil
}

func GetCfgDailySignItem() *[]CfgDailySignItem {
	return cfg.CfgDailySignItems
}

func GetCfgNewUserDailySignItem() *[]CfgDailySignItem {
	return cfg.CfgNewUserDailySignItems
}

type TempProp struct {
	ID               int
	PropID           int
	Name             string
	IsAdd            bool
	IsTowardKnapsack bool
	IsUse            bool
	Expiredat        int64
	Desc             string
}

var PropList = map[int]TempProp{
	values.PropTypeAward: {
		ID:               values.PropTypeAward,
		PropID:           10001,
		Name:             "奖金",
		IsAdd:            true,
		IsTowardKnapsack: false,
		IsUse:            false,
		Expiredat:        -1,
		Desc:             "用户税后奖金超过10元可进行提现申请处理",
	},
	values.PropTypeCoupon: {
		ID:               values.PropTypeCoupon,
		PropID:           10002,
		Name:             "点券",
		IsAdd:            true,
		IsTowardKnapsack: false,
		IsUse:            true,
		Expiredat:        -1,
		Desc:             "用户税后奖金超过10元可进行提现申请处理",
	},
	values.PropTypeCouponFrag: {
		ID:               values.PropTypeCouponFrag,
		PropID:           10003,
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

func SetPropBaseConfig(data map[int]*CfgPropBase) {
	cfg.CfgPropBases = data
	log.Debug("设置缓存成功：%v", cfg.CfgPropBases)
}

func GetPropBaseConfig(propType int) *CfgPropBase {
	data, ok := cfg.CfgPropBases[propType]
	if !ok {
		data := new(CfgPropBase)
		data.ImgUrl = "http://111.230.39.198:10615/download/matchIcon/bg_dianquan.png"
		return data
	}
	return data
}

func GetCfgLianHang() *CfgLianHang {
	return cfg.CfgLianHang
}

func GetCfgNormal() *CfgNormal {
	return cfg.CfgNormal
}
