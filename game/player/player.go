package player

import (
	"ddz/config"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	. "ddz/game/db"
	"ddz/game/rpc"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
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
	SportCenter       SportCenterData // 体总数据
	IsWithdraw        bool
	WithdrawDeadLine  int64
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

// RefreshData 刷新一些数据
func (user *User) RefreshData() {
	one := user.BaseData.UserData
	// 检查是否每日首次登录,是则插入一条登录日志
	t := time.Now().Format("2006-01-02")
	t1, _ := time.Parse("2006-01-02", t)
	today := t1.Unix() - 8*60*60
	if one.Logintime < today {
		game.GetSkeleton().Go(func() {
			// 首次登录插入日志
			db.InsertItemLog(db.ItemLog{
				UID:        user.BaseData.UserData.AccountID,
				Way:        db.FirstLogin,
				Before:     user.BaseData.UserData.Coupon,
				After:      user.BaseData.UserData.Coupon,
				OptType:    db.NormalOpt,
				CreateTime: time.Now().Unix(),
			})
		}, nil)
	}

	// 检查体总数据是否同步,每三小时同步一次
	// if utils.GetZeroTime(time.Unix(one.SportCenter.SyncTime, 0)) != utils.GetZeroTime(time.Now()) && len(one.IDCardNo) > 0 {
	if time.Now().Unix()-one.SportCenter.SyncTime > 3*60*60 && len(one.IDCardNo) > 0 {
		var err error
		sportsCenterData := values.PlayerMasterScoreRet{}
		game.GetSkeleton().Go(func() {
			sportsCenterData, err = edy_api.PlayerMasterScoreQuery(one.IDCardNo)
		}, func() {
			change := false
			if err != nil || len(sportsCenterData.Player_name) == 0 {
				return
			}
			if ret, err := strconv.ParseFloat(sportsCenterData.B, 64); err == nil {
				one.SportCenter.LastBlueScore = one.SportCenter.BlueScore
				one.SportCenter.BlueScore = ret
				change = true
			}
			if ret, err := strconv.ParseFloat(sportsCenterData.R, 64); err == nil {
				one.SportCenter.LastRedScore = one.SportCenter.RedScore
				one.SportCenter.RedScore = ret
				change = true
			}
			if ret, err := strconv.ParseFloat(sportsCenterData.S, 64); err == nil {
				one.SportCenter.LastSilverScore = one.SportCenter.SilverScore
				one.SportCenter.SilverScore = ret
				change = true
			}
			if ret, err := strconv.ParseFloat(sportsCenterData.G, 64); err == nil {
				one.SportCenter.LastGoldScore = one.SportCenter.GoldScore
				one.SportCenter.GoldScore = ret
				change = true
			}
			if change {
				one.SportCenter.LastLevel = one.SportCenter.Level
				one.SportCenter.LastRanking = one.SportCenter.Ranking
				one.SportCenter.Level = sportsCenterData.Level_name
				one.SportCenter.Ranking = sportsCenterData.Ranking
				one.SportCenter.SyncTime = time.Now().Unix()
				game.GetSkeleton().Go(func() {
					UpdateUserData(one.UserID, bson.M{"$set": bson.M{"sportcenter": one.SportCenter}})
				}, nil)
			}
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
	for _, user := range userIDUsers {
		if !user.IsRobot() {
			cnt++
		}
	}
	return cnt
}

func (user *User) SendUserInfo() {
	send := map[string]interface{}{}
	type Sports struct {
		BlueScore     string
		RedScore      string
		SilverScore   string
		GoldScore     string
		Level         string
		Ranking       int // 排名
		TotalCount    int // 总局数
		TotalWin      int // 总胜局
		TotalChampion int // 总冠军数
	}
	sports := Sports{
		BlueScore:     utils.RoundFloat(user.BaseData.UserData.SportCenter.BlueScore, 1),
		RedScore:      utils.RoundFloat(user.BaseData.UserData.SportCenter.RedScore, 1),
		SilverScore:   utils.RoundFloat(user.BaseData.UserData.SportCenter.SilverScore, 1),
		GoldScore:     utils.RoundFloat(user.BaseData.UserData.SportCenter.GoldScore, 1),
		Level:         user.BaseData.UserData.SportCenter.Level,
		Ranking:       user.BaseData.UserData.SportCenter.Ranking,
		TotalChampion: user.BaseData.UserData.SportCenter.TotalChampion,
		TotalCount:    user.BaseData.UserData.SportCenter.TotalCount,
		TotalWin:      user.BaseData.UserData.SportCenter.TotalWin,
	}
	send["Sports"] = sports

	user.WriteMsg(&msg.S2C_UserInfo{
		Info: send,
	})
}

// GetDailyWelfareInfo 获取每日福利详情
func (u *User) GetDailyWelfareInfo() {
	var err error
	reply := &rpc.RPCRet{}
	game.GetSkeleton().Go(func() {
		err = rpc.CallActivityServer("DailyWelfareObj.GetDailyWelfareInfo",
			rpc.RPCGetDailyWelfareInfo{AccountID: u.BaseData.UserData.AccountID}, reply)
	}, func() {
		info := msg.DailyData{}
		defer func() {
			u.WriteMsg(&msg.S2C_GetDailyWelfareInfo{
				Code: reply.Code,
				Desc: reply.Desc,
				Info: info,
			})
		}()
		if err != nil {
			return
		}
		if err := json.Unmarshal([]byte(reply.Data.(string)), &info); err != nil {
			return
		}
		for i := range info.MatchTimeAward {
			propType := values.PropID2Type[info.MatchTimeAward[i].Item]
			info.MatchTimeAward[i].URL = config.GetPropBaseConfig(propType).ImgUrl
		}
		for i := range info.AdditionalAward {
			propType := values.PropID2Type[info.AdditionalAward[i].Item]
			info.AdditionalAward[i].URL = config.GetPropBaseConfig(propType).ImgUrl
		}
	})
}

// DrawDailyWelfare 领取每日福利
func (u *User) DrawDailyWelfare(dailyType, awardIndex int) {
	var err error
	reply := &rpc.RPCRet{}
	game.GetSkeleton().Go(func() {
		err = rpc.CallActivityServer("DailyWelfareObj.DrawDailyWelfare",
			rpc.RPCDrawDailyWelfare{AccountID: u.BaseData.UserData.AccountID, DailyType: dailyType, AwardIndex: awardIndex}, reply)
	}, func() {
		code := reply.Code
		desc := reply.Desc
		defer func() {
			u.WriteMsg(&msg.S2C_GetDailyWelfareInfo{
				Code: code,
				Desc: desc,
			})
		}()
		if err != nil {
			code = 1
			desc = "领取失败!"
			if len(reply.Desc) > 0 {
				desc = reply.Desc
			}
		}
	})
}
