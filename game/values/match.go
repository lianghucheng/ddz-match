package values

import (
	"strconv"
	"strings"
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
	SignIn(uid int)
	SignOut(uid int, matchID string)
	GetNormalConfig() *NormalCofig
	SendMatchDetail(uid int)
	End(matchID string)
	RemoveSignPlayer(uid int)
	CreateOneMatch()
}

// MatchPlayer 比赛玩家对象
type MatchPlayer struct {
	UID        int
	Rank       int
	Nickname   string
	TotalScore int64
	LastScore  int64
	Wins       int
	OpTime     int64
	SignSort   int
	Result     []Result //牌局详细
}

// NormalCofig 需要返回给客户端的通用配置
type NormalCofig struct {
	MatchID          string
	MatchName        string
	MatchType        string // 赛事类型
	MatchDesc        string
	EnterFee         int64
	Award            []string
	AwardDesc        string // 奖励描述
	Recommend        string // 赛事推荐文字信息
	MaxPlayer        int
	AllSignInPlayers []int // 所有已报名该赛事的玩家
	StartTime        int64 // 比赛开始时间或者比赛倒计时
	StartType        int   // 比赛开赛种类
	ReadyTime        int64 // 剩余时间
	Sort             int   // 赛事排序
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
