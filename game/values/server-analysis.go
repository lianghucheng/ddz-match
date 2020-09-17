package values

// LoginLog 玩家登录日志
type LoginLog struct {
	DateTime   int64 `bson:"datetime"`
	UID        int   `bson:"uid"`
	AccountID  int   `bson:"accountid"`
	RecordTime int64 `bson:"recordtime"`
	LoginOrOut int   `bson:"logiorout"` // 1登入,2登出
}
