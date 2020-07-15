package db

// 操作类型
const (
	NormalOpt = iota + 1
	MatchOpt
	ChargeOpt
)

// 定义获取物品的方式
const (
	FirstLogin   = "登录"   // 每日首次登录
	InitPlayer   = "初始赠送" // 初始化用户
	Charge       = "充值"   // 充值
	DailySign    = "签到"   // 签到奖励
	MatchSignIn  = "报名赛事" // 赛事报名
	MatchSignOut = "退出赛事" // 赛事签出
	MatchAward   = "赛事奖励" // 赛事奖励
)

// ItemLog 物品日志
type ItemLog struct {
	UID        int    `bson:"uid"`
	Item       string `bson:"item"`       // 物品名称
	Amount     int64  `bson:"amount"`     // 物品数量
	Way        string `bson:"way"`        // 增加物品的方式
	CreateTime string `bson:"createtime"` // 创建时间
	Before     int64  `bson:"before"`     // 操作前余额
	After      int64  `bson:"after"`      // 操作后余额
	OptType    int    `bson:"opttype"`    // 操作类型
	MatchID    string `bson:"matchid"`    // 赛事id
}
