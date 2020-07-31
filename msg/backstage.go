package msg

import (
	"net/http"
	"sync"
)

func init() {
	Processor.Register(&RPC_AddManagerReq{})
	Processor.Register(&RPC_ShowHall{})
	Processor.Register(&RPC_EditSort{})
	Processor.Register(&RPC_EditMatch{})
	Processor.Register(&RPC_OptMatch{})
	Processor.Register(&RPC_OptUser{})
	Processor.Register(&RPC_ClearInfo{})
	Processor.Register(&RPC_Restart{})
	Processor.Register(&RPC_EditWhiteList{})
	Processor.Register(&RPC_GetOnline{})
}

// RPC_AddManagerReq 后台调用游戏服新增赛事
type RPC_AddManagerReq struct {
	MatchID     string   // 赛事id号
	MatchType   string   // 赛事类型
	MatchName   string   // 赛事名称
	MatchDesc   string   // 赛事描述 `选填`
	Round       int      // 赛制几局
	Card        int      // 赛制几副
	StartType   int      // 比赛开始类型
	StartTime   int64    // 比赛开始时间 `选填`
	LimitPlayer int      // 比赛开始的最少人数
	Recommend   string   // 赛事推荐介绍(在赛事列表界面倒计时左侧的文字信息)
	TotalMatch  int      // 后台配置的该种比赛可创建的比赛次数
	Eliminate   []int    // 每轮淘汰人数 `选填`
	EnterFee    int64    // 报名费
	ShelfTime   int64    // 上架时间
	Sort        int      // 赛事排序
	AwardDesc   string   // 奖励描述 `选填`
	AwardList   string   // 奖励列表
	TablePlayer int      // 一桌的游戏人数 `选填`
	OfficalIDs  []string // 后台配置的可用比赛id号 `选填`
	RoundNum    string   // 几局几副 `选填`

	WG    *sync.WaitGroup     // 用于等待协程返回
	Write http.ResponseWriter // 在协程中返回请求
}

// RPC_ShowHall 后台控制赛事是否在大厅显示
type RPC_ShowHall struct {
	MatchID  string
	ShowHall bool

	WG    *sync.WaitGroup     // 用于等待协程返回
	Write http.ResponseWriter // 在协程中返回请求
}

// RPC_EditSort 后台控制赛事是否在大厅显示
type RPC_EditSort struct {
	MatchID string
	Sort    int

	WG    *sync.WaitGroup     // 用于等待协程返回
	Write http.ResponseWriter // 在协程中返回请求
}

// RPC_EditMatch 后台控制赛事是否在大厅显示
type RPC_EditMatch struct {
	MatchID    string // 赛事id号
	TotalMatch int    // 后台配置的该种比赛可创建的比赛次数
	Eliminate  []int  // 每轮淘汰人数
	EnterFee   int64  // 报名费
	AwardList  string // 奖励列表
	MatchIcon  string // 赛事图标

	WG    *sync.WaitGroup     // 用于等待协程返回
	Write http.ResponseWriter // 在协程中返回请求
}

// RPC_OptMatch 后台控制赛事是否在大厅显示
type RPC_OptMatch struct {
	MatchID string // 赛事id号
	Opt     int    // 操作符，1上架，2下架，3删除

	WG    *sync.WaitGroup
	Write http.ResponseWriter
}

// RPC_OptUser 后台控制玩家封号
type RPC_OptUser struct {
	UID int
	Opt int

	WG    *sync.WaitGroup
	Write http.ResponseWriter
}

// RPC_ClearInfo 后台清除玩家绑定信息
type RPC_ClearInfo struct {
	UID int
	Opt int

	WG    *sync.WaitGroup
	Write http.ResponseWriter
}

// RPC_Restart 后台设置服务器重启
type RPC_Restart struct {
	Config         string `bson:"config"`
	ID             string `bson:"id"`
	TipsTime       int64  `bson:"tipstime"`
	RestartTime    int64  `bson:"restarttime"`
	EndTime        int64  `bson:"endtime"`
	RestartTitle   string `bson:"restarttitle"`
	RestartType    string `bson:"restarttype"`
	Status         int    `bson:"status"`
	RestartContent string `bson:"restartcontent"`
	CreateTime     int64  `bson:"createtime"`

	WG    *sync.WaitGroup
	Write http.ResponseWriter
}

// RPC_EditWhiteList 后台通知服务器白名单有变动
type RPC_EditWhiteList struct {
	WG    *sync.WaitGroup
	Write http.ResponseWriter
}

// RPC_GetOnline 后台获取在线人数
type RPC_GetOnline struct {
	WG    *sync.WaitGroup
	Write http.ResponseWriter
}
