package msg

import (
	"ddz/game/poker"
)

func init() {
	Processor.Register(&S2C_UpdateTotalScore{})
	Processor.Register(&S2C_LandlordMatchRound{})
	Processor.Register(&C2S_LandlordMatchRound{})
}

// 叫分动作
type S2C_ActionLandlordBid struct {
	Position  int
	Countdown int // 倒计时
	Score     []int
}

type C2S_LandlordBid struct {
	Score int //叫的分数
}

type S2C_LandlordBid struct {
	Position int
	Score    int //叫的分数
}

// 加倍动作（只发给自己）
type S2C_ActionLandlordDouble struct {
	Countdown int // 倒计时
}

type C2S_LandlordDouble struct {
	Double bool
}

type S2C_LandlordDouble struct {
	Position int
	Double   bool
}

type S2C_DecideLandlord struct {
	Position int
}

type S2C_UpdateLandlordLastThree struct {
	Cards []int
}

// 出牌动作
type S2C_ActionLandlordDiscard struct {
	ActionDiscardType int // 出牌动作类型
	Position          int
	Countdown         int     // 倒计时
	PrevDiscards      []int   // 上一次出的牌
	Hint              [][]int // 出牌提示
}

type C2S_LandlordDiscard struct {
	Cards []int
}

type S2C_LandlordDiscard struct {
	Position int
	Cards    []int
	CardType int
}

type S2C_ClearAction struct{} // 清除动作

// 单局成绩
type S2C_LandlordRoundResult struct {
	Result       int // 0 失败、1 胜利
	RoomDesc     string
	Spring       bool
	RoundResults []poker.LandlordPlayerRoundResult
	ContinueGame bool // 是否继续游戏
	Type         int  // 0 防守方 1 进攻方
	Position     int
	Process      []string //总进度
	Allcount     int      //总局数
	RankOrder    int      //排名
	CurrCount    int      //当前局数
	Countdown    int      //下一局等待时间
}

type S2C_GameStart struct{}

type S2C_UpdatePokerHands struct {
	Position      int
	Hands         []int // 手牌
	NumberOfHands int   // 手牌数量
}

// 系统托管
type C2S_SystemHost struct {
	Host bool
}

type S2C_SystemHost struct {
	Position int
	Host     bool
}

//对局积分
type S2C_UpdateTotalScore struct {
	Result []Result
}

type Result struct {
	Position   int
	TotalScore int64
}

//获取对局排名
type C2S_LandlordMatchRound struct {
}

type S2C_LandlordMatchRound struct {
	RoundResults []poker.LandlordRankData
}

// C2S_GetMatchList 获取赛事列表
type C2S_GetMatchList struct {
}

// S2C_GetMatchList 返回赛事列表
type S2C_GetMatchList struct {
	List []OneMatch
}

type OneMatch struct {
	MatchID   string
	MatchName string
	SignInNum int
	Recommend string
	MaxPlayer int
	EnterFee  int64
	IsSign    bool
}
