package values

import (
	"ddz/msg"
	"strconv"
	"strings"

	"github.com/name5566/leaf/timer"
)

// 赛事配置
var (
	MatchTypeConfig = map[string]msg.OneMatchType{} // 赛事类型配置
)

// Match 比赛接口
type Match interface {
	SignIn(uid int) error
	SignOut(uid int) error
	CheckStart() // 判断比赛是否开始
	Start()
	SplitTable()             // 分桌逻辑
	RoundOver(roomID string) // 单局结束，获取结果
	End()
	GetRank(uid int)         // 获取排名情况
	SendRoundResult(uid int) // 给玩家发送单局结算
	SendFinalResult(uid int) // 给玩家发送总结算
}

// MatchManager 比赛配置接口
type MatchManager interface {
	NewManager()
	SignIn(uid int)
	SignOut(uid int, matchID string)
	GetNormalConfig() *NormalCofig
	SetNormalConfig(config *NormalCofig)
	SendMatchDetail(uid int)
	End(matchID string)
	RemoveSignPlayer(uid int)
	CreateOneMatch()
	Save() error
	CheckNewConfig()
	ClearLastMatch()
}

// MatchPlayer 比赛玩家对象
type MatchPlayer struct {
	UID        int
	Rank       int
	Nickname   string
	TotalScore int64
	LastScore  int64
	Wins       int
	OneOpTime  int64 // 单局操作时间
	OpTime     int64
	SignSort   int
	Result     []Result //牌局详细
	Multiples  string   // 当局所有加倍详z
}

// NormalCofig 需要返回给客户端的通用配置
type NormalCofig struct {
	MatchID          string
	MatchName        string
	MatchType        string // 赛事类型
	MatchDesc        string
	EnterFee         int64
	State            int
	Award            []string
	AwardDesc        string // 奖励描述
	Recommend        string // 赛事推荐文字信息
	MaxPlayer        int
	AllSignInPlayers []int        // 所有已报名该赛事的玩家
	StartTime        int64        // 比赛开始时间或者比赛倒计时
	StartType        int          // 比赛开赛种类
	ReadyTime        int64        // 剩余时间
	Sort             int          // 赛事排序
	ShowHall         bool         // 首页展示
	MatchIcon        string       // 赛事图标
	SonMatchID       string       // 自赛事id
	TotalMatch       int          // 总赛事场次
	Eliminate        []int        // 淘汰人数
	AwardList        string       // 奖励
	StartTimer       *timer.Timer // 上架倒计时
}

// MatchRecord 记录一局比赛所有玩家的手牌，输赢信息等
type MatchRecord struct {
	RoundCount int    // 第几局
	CardCount  int    // 第几副牌
	RoomCount  int    // 房间编号
	UID        int    // 用户id
	Identity   int    //0 防守方 1 进攻方
	Name       string // 玩家姓名
	HandCards  []int  //手牌
	ThreeCards []int  //底牌
	Event      int    //0:失败 1:胜利
	Score      int64  //得分
	Multiples  string //倍数
}

// UserMatchReview 用户后台的赛事列表总览
type UserMatchReview struct {
	UID            int
	AccountID      int
	MatchID        string
	MatchType      string
	MatchName      string
	MatchTotal     int
	MatchWins      int
	MatchFails     int
	AverageBatting int
	Coupon         int64
	AwardMoney     int64
	PersonalProfit int64
}

// MatchData 统计一些玩家的赛事数据
type MatchData struct {
	TotalCount int   // 总局数
	WeekCount  int   // 周局数
	MonthCount int   // 月局数
	RecordTime int64 // 记录时间
}

// GetAwardType 获取奖励类型
func GetAwardType(award string) string {
	if strings.Index(award, Money) != -1 {
		return Money
	}
	if strings.Index(award, Coupon) != -1 {
		return Coupon
	}
	return Unknown
}

// ParseAward 解析奖励的数量
func ParseAward(award string) float64 {
	// log.Debug("parse award:%v", award)
	num := []byte{}
	for _, s := range []byte(award) {
		if s <= 57 && s >= 46 {
			num = append(num, s)
		}
	}
	b, _ := strconv.ParseFloat(string(num), 64)
	return b
}

// GetMoneyAward 获取奖励字段中的奖金之和
func GetMoneyAward(award string) float64 {
	var amount float64
	s := strings.Split(award, ",")
	for _, one := range s {
		if GetAwardType(one) == Money {
			amount += ParseAward(one)
		}
	}
	return amount
}
