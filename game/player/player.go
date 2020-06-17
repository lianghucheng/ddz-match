package player

import (
	. "ddz/game/db"
	"fmt"
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

type UserData struct {
	UserID            int `bson:"_id"`
	AccountID         int
	Nickname          string
	Headimgurl        string
	Sex               int // 1 男性，2 女性
	LoginIP           string
	Token             string
	ExpireAt          int64 // token 过期时间
	Role              int   // 1 玩家、2 代理、3 管理员、4 超管
	Username          string
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
}

type User struct {
	gate.Agent
	State          int
	BaseData       *BaseData
	HeartbeatTimer *timer.Timer
	HeartbeatStop  bool
}
type BaseData struct {
	UserData *UserData
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
