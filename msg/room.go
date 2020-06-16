package msg

func init() {
	Processor.Register(&S2C_RoomPanel{})
	Processor.Register(&S2C_MineRoundRank{})
	Processor.Register(&C2S_GetGameRecord{})
	Processor.Register(&S2C_GetGameRecord{})
}

const (
	S2C_EnterRoom_OK = 0
)

type S2C_EnterRoom struct {
	Error      int
	Position   int
	BaseScore  int
	MaxPlayers int // 最大玩家数
}

type C2S_GetAllPlayers struct{}

type S2C_SitDown struct {
	Position   int
	AccountID  int
	LoginIP    string
	Nickname   string
	Headimgurl string
	Sex        int
	Chips      int
}
type S2C_RoomPanel struct {
	Spring      int //春天  0 显示-
	LSpring     int //反春天
	Boom        int //炸弹数量
	BaseScore   int //底分
	DealerScore int //叫分
	Ming        int //明牌
	Public      int //公共
	Dealer      int //庄家
	Xian        int //防守方
	Total       int //总倍数
}

type S2C_MineRoundRank struct {
	Result    int // 0 失败、1 胜利
	RankOrder int
	Award     float64
	Spring    bool
	Type      int // 0 防守方 1 进攻方
}

// 获取战绩记录
type C2S_GetGameRecord struct {
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
	MatchType  int //1 海选赛 2 复式赛等等
}

// 战绩记录

type S2C_GetGameRecord struct {
	Items      []GameRecord //记录数据
	Total      int          //记录数量
	PageNumber int          //当前页
	PageSize   int          //一页显示的条数
}
type GameRecord struct {
	UserId    int          //用户ID
	MatchId   string       //赛事ID
	MatchType string       //赛事类型
	Desc      string       //赛事
	Level     int          //名次
	Award     float64      //奖励
	Count     int          //完成局数
	Total     int64        //总得分
	Last      int64        //尾副得分
	Wins      int          //获胜次数
	Period    int64        //累计时长
	Rank      []Rank       //排行
	Result    []GameResult //牌局详细
	CreateDat int64        //时间
}

type Rank struct {
	Level    int     //名次
	NickName string  //用户名
	Count    int     //完成局数
	Total    int64   //总得分
	Last     int64   //尾副牌得分
	Wins     int     //获胜次数
	Period   int64   //累计时长
	Sort     int     //报名次序
	Award    float64 //奖励

}

type GameResult struct {
	Count      int   //第一局
	Event      int   //0:失败 1:胜利
	Identity   int   //0 防守方 1 进攻方
	Bottom     int   //底分
	Multiple   int   //倍数
	Score      int64 //得分
	HandCards  []int //手牌
	ThreeCards []int //底牌

}
