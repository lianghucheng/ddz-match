package msg

func init() {
	Processor.Register(&C2S_Knapsack{})
	Processor.Register(&S2C_Knapsack{})
	Processor.Register(&C2S_UseProp{})
	Processor.Register(&S2C_UseProp{})
	Processor.Register(&C2S_UserInfo{})
	Processor.Register(&S2C_UserInfo{})
	Processor.Register(&C2S_GetDailyWelfareInfo{})
	Processor.Register(&S2C_GetDailyWelfareInfo{})
	Processor.Register(&C2S_DrawDailyWelfareInfo{})
	Processor.Register(&S2C_DrawDailyWelfareInfo{})
	Processor.Register(&S2C_HorseRaceLamp{})
	Processor.Register(&S2C_Notice{})
	Processor.Register(&S2C_Activity{})
	Processor.Register(&C2S_ActivityClick{})
	Processor.Register(&S2C_NewPoint{})
}

type C2S_Knapsack struct {
}

type KnapsackProp struct {
	PropID    int    //道具id
	Name      string //名称
	Num       int    //数量
	IsUse     bool   //是否可使用
	Expiredat int64  //过期时间，-1表示永久
	Desc      string //描述
	Imgurl    string //
	Createdat int64  //创建时间
}

type S2C_Knapsack struct {
	Props []KnapsackProp //道具数据列表
}

type C2S_UseProp struct {
	PropID int //道具id
	Amount int //要领取的结果数量，比如40个碎片换取2个点券，此时传输2
}

type S2C_UseProp struct {
	Error  int    //错误码
	ErrMsg string //错误信息
	Name   string
	PropID int
	Amount int
	Imgurl string
}

type C2S_UserInfo struct {
	AccountID int // 玩家id
}

type S2C_UserInfo struct {
	Info interface{}
}

// C2S_GetDailyWelfareInfo 获取每日福利详情
type C2S_GetDailyWelfareInfo struct {
}

// S2C_GetDailyWelfareInfo 获取每日福利详情
type S2C_GetDailyWelfareInfo struct {
	Code int
	Desc string
	Info DailyData
}

// C2S_DrawDailyWelfareInfo 领取每日福利
type C2S_DrawDailyWelfareInfo struct {
	DailyType  int // 奖励类型
	AwardIndex int // 领取奖励序列号
}

// S2C_DrawDailyWelfareInfo 领取每日福利
type S2C_DrawDailyWelfareInfo struct {
	Code int
	Desc string
	Name   string
	PropID int
	Amount float64
	ImgUrl string
}

// DailyData 玩家每日数据
type DailyData struct {
	MatchTime       int64          // 参赛时间
	MatchCount      int64          // 参赛次数
	MatchTimeAward  []OneItemAward `bson:"MatchTimeAward"`  // 参赛时长奖励
	AdditionalAward []OneItemAward `bson:"AdditionalAward"` // 额外奖励
}

// OneItemAward 单个目标奖励对象
type OneItemAward struct {
	Item         int    `bson:"Item"` // 物品ID
	URL          string // 物品图片地址
	AwardAmount  int    `bson:"AwardAmount"`  // 奖励数量
	TargetAmount int64  `bson:"TargetAmount"` // 达成条件
	Status       int    `bson:"Status"`       // 领取状态1未完成,2已完成未领取,3已领取
}

type HorseRaceLamp struct {
	UserName  string
	MatchName string
	Amount    float64
}

type S2C_HorseRaceLamp struct {
	Template string
	//Info []map[string]string
	LinkMatchID string
}

type ActivityMsg struct {
	ID      int
	Order   int    //排序
	Title   string //活动标题
	Img     string //图片
	Matchid string //关联赛事id
	Link    string //活动连接
}

type S2C_Activity struct {
	Datas []ActivityMsg
}

type NoticeMsg struct {
	ID          int    `bson:"_id"` //唯一标识
	Order       int    //排序
	ColTitle    string //栏目标题
	NoticeTitle string //公告标题
	Content     string //公告内容
	Signature   string //公告落款
	Img         string //公告图片
}

type S2C_Notice struct {
	Datas []NoticeMsg
}

type C2S_ActivityClick struct {
	ID int
}

type S2C_NewPoint struct {
	Datas []map[int]bool
}
