package values

const (
	NewPointIconNotice    = 1 //公告
	NewPointIconActivity  = 2 //活动
	NewPointIconUserInfo  = 3 //我的
	NewPointIconDailySign = 4 //签到
	NewPointIconWelfare   = 5 //福利
	NewPointIconMail      = 6 //邮件
	NewPointIconTask      = 7 //任务
	NewPointIconKnapsack  = 8 //背包
)

type ActivityControl struct {
	ID           int    `bson:"_id"` //唯一标识
	Order        int    //排序
	Title        string //活动标题
	Img          string //图片
	Matchid      string //关联赛事id
	Link         string //活动连接
	Status       int    //状态，0是下架，1是上架
	PrevUpedAt   int    //上架时间
	PrevDownedAt int    //下架时间
	ClickCnt     int    //点击量

	CreatedAt int
	UpdatedAt int
	DeletedAt int
}

type NoticeControl struct {
	ID           int    `bson:"_id"` //唯一标识
	Order        int    //排序
	ColTitle     string //栏目标题
	NoticeTitle  string //公告标题
	Status       int    //状态，0是下架，1是上架
	PrevUpedAt   int    //上架时间戳
	PrevDownedAt int    //下架时间戳
	Operator     string //操作人
	Content      string //公告内容
	Signature    string //公告落款

	CreatedAt int
	UpdatedAt int
	DeletedAt int
}
