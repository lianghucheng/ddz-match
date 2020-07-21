package msg

func init() {
	Processor.Register(&C2S_Apply{})
	Processor.Register(&S2C_Apply{})
	Processor.Register(&S2C_MatchPrepare{})
	Processor.Register(&S2C_MatchNum{})
	Processor.Register(&S2C_MatchInfo{})
}

//告诉玩家参加的赛事即将开赛
type S2C_MatchPrepare struct {
	MatchId string
}
type C2S_Apply struct {
	MatchId string //赛事ID
	Action  int    //1:报名 2:取消报名
}

const (
	S2C_Error_MatchId  = 1 //赛事不存在
	S2C_Error_Coupon   = 2 //点券不足
	S2C_Error_Action   = 3 //已报名(等待开赛)
	S2C_Error_Match    = 4 //玩家已报名了其它赛事
	S2C_Error_Realname = 5 //玩家未实名
	S2C_Error_MoreRobot = 6 //机器人过量
)

type S2C_Apply struct {
	Error  int
	RaceID string
	Action int
	Count  int //当前赛事人数
}
type S2C_MatchInfo struct {
	RoundNum       string //赛制 两局一副
	Process        string //进程 第2局 第1幅
	Level          string //排名 1/3
	Competition    string //晋级 前3晋级
	AwardList      string // 奖励列表
	MatchName      string //比赛名称
	Duration       int64  //时长
	WinCnt         int    //获胜次数
	AwardPersonCnt int    //奖励人数
}

type S2C_MatchNum struct {
	MatchId string
	Count   int //已报名人数
}
