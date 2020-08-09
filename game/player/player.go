package player

import (
	"ddz/game/db"
	. "ddz/game/db"
	"ddz/game/values"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/timer"
	"github.com/name5566/leaf/util"
	"github.com/szxby/tools/log"
)

const (
	DefaultAvatar = "https://www.shenzhouxing.com/ruijin/dl/img/avatar.jpg"
	RoleRobot     = -2 // 机器人
	RoleBlack     = -1 // 黑名单
	RolePlayer    = 1  // 玩家
	RoleAgent     = 2  // 代理
	RoleAdmin     = 3  // 管理员
	RoleRoot      = 4  // 最高管理员
)

// 保存所有玩家信息
var (
	UserIDUsers = make(map[int]*User)
)

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
	LastTakenMail     int64
	RealName          string
	IDCardNo          string
	BankCardNo        string
	SetNickNameCount  int
	TakenFee          float64
	FirstLogin        bool
		IsWithdraw   	  bool
		WithdrawDeadLine int64
}

type User struct {
	gate.Agent
	State          int
	BaseData       *BaseData
	HeartbeatTimer *timer.Timer
	HeartbeatStop  bool
}
type BaseData struct {
	UserData    *UserData
	MatchPlayer *values.MatchPlayer
}
type AgentInfo struct {
	User *User
}

func (u *User) GetUserData() *UserData {
	return u.BaseData.UserData
}

func (data *UserData) InitValue(channel int) error {
	userID, err := MongoDBNextSeq("users")
	if err != nil {
		return fmt.Errorf("get next users id error: %v", err)
	}
	data.UserID = userID
	data.Role = RolePlayer
	data.AccountID = time.Now().Year()*100 + userID
	data.CreatedAt = time.Now().Unix()
	data.Channel = channel
	data.Nickname = "用户" + strconv.Itoa(data.AccountID)
	// 初始化点券
	// data.Coupon = 5
	// db.InsertItemLog(userID, data.Coupon, values.Coupon, db.InitPlayer)
	data.FirstLogin = true
	return nil
}

func SaveUserData(userdata *UserData) {
	data := util.DeepClone(userdata)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	id := data.(*UserData).UserID
	_, err := db.DB(DB).C("users").UpsertId(id, data)
	if err != nil {
		log.Error("save user %v data error: %v", id, err)
	}
}

func UpdateUserData(id int, update interface{}) {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	_, err := db.DB(DB).C("users").UpsertId(id, update)
	if err != nil {
		log.Error("update user %v data error: %v", id, err)
	}

}

func ReadUserDataByID(id int) *UserData {
	userData := new(UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)

	err := db.DB(DB).C("users").Find(bson.M{"_id": id}).One(userData)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return userData
}

func ReadUserDataByAid(accountid int) *UserData {
	userData := new(UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	err := db.DB(DB).C("users").Find(bson.M{"accountid": accountid}).One(userData)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return userData
}

func (user *User) UID() int {
	return user.BaseData.UserData.UserID
}

func (user *User) RealName() string {
	return user.BaseData.UserData.RealName
}

func (user *User) PhoneNum() string {
	return user.BaseData.UserData.Username
}

func (user *User) Fee() float64 {
	return user.BaseData.UserData.Fee
}

func (user *User) Coupon() int64 {
	return user.BaseData.UserData.Coupon
}

func (user *User) IDCardNo() string {
	return user.BaseData.UserData.IDCardNo
}

func (user *User) BankCardNo() string {
	return user.BaseData.UserData.BankCardNo
}

func (user *User) AcountID() int {
	return user.BaseData.UserData.AccountID
}

// CheckFirstLogin 检查是否每日的首次登录
func (user *User) CheckFirstLogin() {
	// var oneDay int64 = 24 * 60 * 60
	t := time.Now().Format("2006-01-02")
	t1, _ := time.Parse("2006-01-02", t)
	today := t1.Unix() - 8*60*60
	// log.Debug("today:%v,logintime:%v", today, user.BaseData.UserData.Logintime)
	if user.BaseData.UserData.Logintime < today {
		// 首次登录插入日志
		db.InsertItemLog(db.ItemLog{
			UID:        user.BaseData.UserData.AccountID,
			Way:        db.FirstLogin,
			Before:     user.BaseData.UserData.Coupon,
			After:      user.BaseData.UserData.Coupon,
			OptType:    db.NormalOpt,
			CreateTime: time.Now().Unix(),
		})
	}
}

func (user *User) IsTest() bool {
	if user.GetUserData().Username[:4] == "test" {
		return true
	}
	return false
}

func (user *User) IsRobot() bool {
	return user.GetUserData().Role == RoleRobot
}

func CalcOnlineCnt(userIDUsers map[int]*User) int {
	cnt := 0
	for _,user := range userIDUsers {
		if !user.IsRobot() {
			cnt++
		}
	}
	return cnt
}