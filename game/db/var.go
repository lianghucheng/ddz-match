package db

// 定义获取物品的方式
const (
	Charge       = "charge"       // 充值
	DailySign    = "sign"         // 签到奖励
	MatchSignIn  = "matchSignIn"  // 赛事报名
	MatchSignOut = "matchSignOut" // 赛事签出
	MatchAward   = "matchAward"   // 赛事奖励
)

// ItemLog 物品日志
type ItemLog struct {
	UID        int    `bson:"uid"`
	Item       string `bson:"item"`       // 物品名称
	Amount     int64  `bson:"amount"`     // 物品数量
	Way        string `bson:"way"`        // 增加物品的方式
	CreateTime string `bson:"createtime"` // 创建时间
}
