package values

type DDZGameRecord struct {
	UserId    int      //用户ID
	MatchId   string   //赛事ID
	MatchType string   //赛事类型
	Desc      string   //赛事
	Level     int      //名次
	Award     float64  //奖励
	Count     int      //完成局数
	Total     int64    //总得分
	Last      int64    //尾副得分
	Wins      int      //获胜次数
	Period    int64    //累计时长
	Rank      []Rank   //排行
	Result    []Result //牌局详细
	CreateDat int64    //时间
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

type Result struct {
	Count      int   // 第一局
	CardCount  int   // 第几副牌
	Event      int   //0:失败 1:胜利
	Identity   int   //0 防守方 1 进攻方
	Bottom     int   //底分
	Multiple   int   //倍数
	Score      int64 //得分
	HandCards  []int //手牌
	ThreeCards []int //底牌
}