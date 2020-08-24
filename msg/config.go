package msg

func init() {
	Processor.Register(&C2S_RaceInfo{})
	Processor.Register(&S2C_RaceInfo{})
	Processor.Register(&C2S_RaceDetail{})
	Processor.Register(&S2C_RaceDetail{})
}

type RaceInfo struct {
	ID           string  //赛事Id
	Desc         string  //赛事名称
	Award        float64 //赛事
	EnterFee     float64 //报名费
	ConDes       string  //赛事开赛条件
	JoinNum      int     //赛事报名人数
	AllPlayerNum int     // 赛事参与总人数
	StartTime    int64   // 比赛开始时间
	StartType    int     // 比赛开赛方式
	IsSign       bool    // 是否报名
	MatchType    string
	Eliminate    []int // 晋级条件
}

type C2S_RaceInfo struct {
}

type S2C_RaceInfo struct {
	Races []OneMatch
}

type C2S_RaceDetail struct {
	ID string
}

type S2C_RaceDetail struct {
	ID            string  //赛事ID
	Desc          string  //
	AwardDesc     string  //奖励描述
	AwardList     string  // 奖励别表
	MatchType     string  //赛事类型
	RoundNum      string  //对局副数
	StartTime     int64   // 比赛开始时间
	StartType     int     // 比赛开赛方式
	ConDes        string  //赛事开赛条件
	SignNum       int     //报名人数
	SignNumDetail bool    //当前报名数是否可点击
	EnterFee      float64 //报名费
	IsSign        bool    //报名按钮的状态(报名,取消)
	Eliminate     []int   // 晋级条件
}
