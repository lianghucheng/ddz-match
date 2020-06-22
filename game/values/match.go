package values

import (
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
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
	GetRank(uid int) // 获取排名情况
}

// MatchManager 比赛配置接口
type MatchManager interface {
	SignIn(uid int)
	SignOut(uid int)
	GetNormalConfig() NormalCofig
	SendMatchDetail(uid int)
	End(matchID string)
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
	log.Debug("parse award:%v", award)
	num := []byte{}
	for _, s := range []byte(award) {
		if s <= 57 && s >= 46 {
			num = append(num, s)
		}
	}
	b, _ := strconv.ParseFloat(string(num), 64)
	return b
}
