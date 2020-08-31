package edy_api

import (
	"ddz/edy_api/internal"
	"ddz/edy_api/internal/base"
	"ddz/game/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

func RealAuthApi(accountid int, idCardNo, realName, phoneNum string) error {
	idBind := internal.NewIDBindReq(accountid, idCardNo, realName, phoneNum)
	return idBind.IdCardBind()
}

func RealAuthApi2(accountid int, idCardNo, realName, phoneNum string) error {
	return nil
}

func BandBankCardAPI(accountid int, bankNo, BankName, BankAccount string) error {
	bindBankCard := internal.NewBindBankCardReq(accountid, bankNo, BankName, BankAccount)
	return bindBankCard.BindBankCard()
}

func BandBankCardAPI2(accountid int, bankNo, BankName, BankAccount string) error {
	return nil
}

type SportCenterData struct {
	TotalCount      int     // 总局数
	TotalWin        int     // 总胜局
	TotalChampion   int     // 总冠军数
	BlueScore       float64 // 蓝分
	RedScore        float64 // 红分
	SilverScore     float64 // 银分
	GoldScore       float64 // 金分
	Level           string  // 等级称号
	Ranking         int     // 排名
	SyncTime        int64   // 与体总同步时间
	LastBlueScore   float64 // 同步前蓝分
	LastRedScore    float64 // 同步前红分
	LastSilverScore float64 // 同步前银分
	LastGoldScore   float64 // 同步前金分
	LastLevel       string  // 同步前等级称号
	LastRanking     int     // 同步前排名
	WalletStatus    int     // 钱包状态 0锁定,1正常
}

type UserData struct {
	UserID            int `bson:"_id"`
	AccountID         int
	Nickname          string
	Headimgurl        string
	Sex               int // 1 男性，2 女性
	LoginIP           string
	Logintime         int64 // 登录时间
	Token             string
	ExpireAt          int64 // token 过期时间
	Role              int   // 1 玩家、2 代理、3 管理员、4 超管
	Username          string
	Password          string
	Coupon            int64 // 点券
	Wins              int   // 胜场
	CreatedAt         int64
	UpdatedAt         int64
	PlayTimes         int     //当天对局次数
	Online            bool    //玩家是否在线
	Channel           int     //渠道号。0：圈圈   1：搜狗   2:IOS
	Fee               float64 //税后余额
	SignTimes         int
	DailySign         bool
	DailySignDeadLine int64
	NotNewDailySign   bool
	LastTakenMail     int64
	RealName          string
	IDCardNo          string
	BankCardNo        string
	SetNickNameCount  int
	TakenFee          float64
	FirstLogin        bool
	SportCenter       SportCenterData // 体总数据
	IsWithdraw        bool
	WithdrawDeadLine  int64
}

func ReadUserDataByID(id int) *UserData {
	userData := new(UserData)
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	err := se.DB(db.DB).C("users").Find(bson.M{"_id": id}).One(userData)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return userData
}

func WithDrawAPI(userid int, amount float64) error {
	ud := ReadUserDataByID(userid)
	req := new(PlayerCashoutReq)
	req.Cp_id = base.CpID
	req.Player_id = fmt.Sprintf("%v", ud.AccountID)
	req.Player_id_number = ud.IDCardNo
	data, err := PlayerCashout(*req)
	if err != nil {
		return errors.New(data["resp_msg"].(string))
	}
	return nil
}

type PlayerCashoutReq struct {
	Cp_id            string `json:"cp_id"`
	Player_id        string `json:"player_id"`
	Player_id_number string `json:"player_id_number"`
}

// PlayerCashout 玩家提现
func PlayerCashout(data PlayerCashoutReq) (map[string]interface{}, error) {
	data.Cp_id = base.CpID
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/player/bonous/withdraw", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	msg := map[string]interface{}{}
	if err := json.Unmarshal(ret, &msg); err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	err = checkCode(ret)
	return msg, err
}
