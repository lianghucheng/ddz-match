package msg

func init() {
	Processor.Register(&C2S_EndMatch{})
	Processor.Register(&RPC_SendMail{})
	Processor.Register(&RPC_SendRaceInfo{})
	Processor.Register(&RPC_WriteAwardFlowData{})
	Processor.Register(&RPC_SendMatchEndMail{})
	Processor.Register(&RPC_SendInterruptMail{})
	Processor.Register(&RPC_NotifyPayOK{})
	Processor.Register(&RPC_AddFee{})
	Processor.Register(&RPC_AddAward{})
	Processor.Register(&RPC_UpdateAwardInfo{})
	Processor.Register(&RPC_UpdateCoupon{})
	Processor.Register(&RPC_UpdateHeadImg{})
	Processor.Register(&RPC_SendKnapsack{})
	Processor.Register(&RPC_AddCouponFrag{})
	Processor.Register(&RPC_SendPayAccount{})
	Processor.Register(&RPC_UpdateBankCardNo{})
	Processor.Register(&RPC_HorseLamp{})
	Processor.Register(&RPC_ActivityNotify{})
	Processor.Register(&RPC_NoticeNotify{})
	Processor.Register(&RPC_RealnameUpdate{})
}

type C2S_EndMatch struct {
	MatchId string //赛事ID
	Id      int    //用户ID
}

type RPC_SendMail struct {
	ID int //用户ID
}

type RPC_SendMatchEndMail struct {
	Userid    int
	MatchName string
	Order     int
	Award     float64
}

type RPC_SendInterruptMail struct {
	Userid    int
	MatchName string
	Coupon    int64
}

type RPC_SendRaceInfo struct {
	ID int //Userid
}

type RPC_WriteAwardFlowData struct {
	Userid  int
	Amount  float64
	Matchid string
}

type RPC_NotifyPayOK struct {
	TotalFee  int
	AccountID int
	Amount    int
}

type RPC_AddFee struct {
	FeeType string  `json:"fee_type"`
	Userid  int     `json:"userid"`
	Amount  float64 `json:"amount"`
}

type RPC_AddAward struct {
	Secret string
	Uid    int
	Amount float64
}

type RPC_UpdateAwardInfo struct {
	Uid int
}

type RPC_UpdateCoupon struct {
	Secret    string
	Accountid int
	Amount    int
}

type RPC_UpdateHeadImg struct {
	Secret    string
	Accountid int
	HeadImg   string
}

type RPC_SendKnapsack struct {
	Aid int
}

type RPC_AddCouponFrag struct {
	Secret    string
	Accountid int
	Amount    int
}

type RPC_SendPayAccount struct {
}

type RPC_UpdateBankCardNo struct {
	Userid     int
	BankCardNo string
}

type RPC_HorseLamp struct {
	ID           int
	Template     string
	ExpiredAt    int    //过期时间戳
	TakeEffectAt int    //发布时间戳
	Duration     int    //间隔时长，单位s
	LinkMatchID  string //关联赛事id
	Level        int
	NextTmp      int
}

type RPC_ActivityNotify struct {

}

type RPC_NoticeNotify struct {

}

type RPC_RealnameUpdate struct {
	Accountid int //用户id
	RealName string //姓名
	IDCardNo string //身份证号
}