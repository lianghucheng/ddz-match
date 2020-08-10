package msg

import (
	"github.com/name5566/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {

	Processor.Register(&C2S_TokenLogin{})
	Processor.Register(&C2S_AccountLogin{})
	Processor.Register(&C2S_Heartbeat{})

	Processor.Register(&C2S_GetAllPlayers{})

	Processor.Register(&C2S_LandlordBid{})

	Processor.Register(&C2S_LandlordDouble{})

	Processor.Register(&C2S_LandlordDiscard{})

	Processor.Register(&C2S_SystemHost{})

	Processor.Register(&S2C_Close{})
	Processor.Register(&S2C_Login{})
	Processor.Register(&S2C_Heartbeat{})

	Processor.Register(&S2C_EnterRoom{})
	Processor.Register(&S2C_SitDown{})

	Processor.Register(&S2C_GameStart{})
	Processor.Register(&S2C_UpdatePokerHands{})
	Processor.Register(&S2C_ActionLandlordBid{})
	Processor.Register(&S2C_LandlordBid{})
	Processor.Register(&S2C_DecideLandlord{})
	Processor.Register(&S2C_UpdateLandlordLastThree{})
	Processor.Register(&S2C_ActionLandlordDouble{})
	Processor.Register(&S2C_LandlordDouble{})

	Processor.Register(&S2C_ActionLandlordDiscard{})
	Processor.Register(&S2C_LandlordDiscard{})
	Processor.Register(&S2C_ClearAction{})
	Processor.Register(&S2C_LandlordRoundResult{})

	Processor.Register(&S2C_SystemHost{})

	Processor.Register(&C2S_DailySign{})
	Processor.Register(&S2C_DailySign{})
	Processor.Register(&S2C_DailySignItems{})

	Processor.Register(&C2S_FeedBack{})
	Processor.Register(&S2C_FeedBack{})
	Processor.Register(&S2C_SendMail{})
	Processor.Register(&C2S_ReadMail{})
	Processor.Register(&C2S_DeleteMail{})
	Processor.Register(&S2C_DeleteMail{})
	Processor.Register(&C2S_TakenMailAnnex{})
	Processor.Register(&S2C_TakenMailAnnex{})
	Processor.Register(&C2S_RankingList{})
	Processor.Register(&S2C_RankingList{})
	Processor.Register(&C2S_RealNameAuth{})
	Processor.Register(&S2C_RealNameAuth{})
	Processor.Register(&C2S_BindBankCard{})
	Processor.Register(&S2C_BindBankCard{})
	Processor.Register(&S2C_AwardInfo{})
	Processor.Register(&C2S_WithDraw{})
	Processor.Register(&S2C_WithDraw{})
	Processor.Register(&C2S_GetMatchList{})
	Processor.Register(&S2C_GetMatchList{})

	Processor.Register(&C2S_GetGameRecord{})
	Processor.Register(&S2C_GetGameRecord{})
	Processor.Register(&C2S_GetGameRankRecord{})
	Processor.Register(&S2C_GetGameRankRecord{})
	Processor.Register(&C2S_GetGameResultRecord{})
	Processor.Register(&S2C_GetGameResultRecord{})
	Processor.Register(&S2C_PayOK{})
	Processor.Register(&S2C_LandlordRoundFinalResult{})
	Processor.Register(&C2S_GetGameRecordAll{})
	Processor.Register(&S2C_GetGameRecordAll{})
	Processor.Register(&C2S_CreateEdyOrder{})
	Processor.Register(&S2C_CreateEdyOrder{})
	Processor.Register(&S2C_PriceMenu{})
	Processor.Register(&S2C_OnlineUserNum{})
}

type C2S_Heartbeat struct{}

type S2C_Heartbeat struct{}

type C2S_DailySign struct {
}

type S2C_DailySign struct {
	Name   string
	PropID int
	Amount float64
}

type S2C_DailySignItems struct {
	SignItems []DailySignItems
	IsSign    bool //今日是否已签到
}

const (
	SignFinish = 1
	SignAccess = 2
	SignDeny   = 3
)

type DailySignItems struct {
	Name   string
	PropID int
	Amount float64
	Status int
}

type C2S_FeedBack struct {
	Title   string
	Content string
}

const (
	S2C_FeedBaock_OK   = 1
	S2C_FeedBaock_Fail = 2
)

type S2C_FeedBack struct {
	Error int
}

type Annex struct {
	Type int
	Num  int
	Desc string
}

type UserMail struct {
	ID        int64   `bson:"_id"` //唯一主键
	CreatedAt int64   //收件时间
	Title     string  //主题
	Content   string  //内容
	Annexes   []Annex //附件
	Status    int64   //邮件状态
}

type S2C_SendMail struct {
	Datas []UserMail
}

type C2S_ReadMail struct {
	ID int64 //唯一主键
}

type C2S_DeleteMail struct {
	ID int64 //唯一主键
}

const (
	S2C_DeleteMail_OK   = 0
	S2C_DeleteMail_Fail = 1
)

type S2C_DeleteMail struct {
	Error int
}

type C2S_TakenMailAnnex struct {
	ID int64 //唯一主键
}

const (
	S2C_TakenMailAnnex_OK   = 0
	S2C_TakenMailAnnex_Fail = 1
)

type S2C_TakenMailAnnex struct {
	Error   int
	Annexes []Annex //附件
}

type C2S_RankingList struct {
}

type Ranking struct {
	Order    int
	NickName string
	Value    string
}

type RankByType struct {
	Name  string
	Ranks []Ranking
	Mine  Ranking
}

type S2C_RankingList struct {
	ChipRankingList  []RankByType
	AwardRankingList []RankByType
}

type C2S_RealNameAuth struct {
	RealName string
	IDCardNo string
}

const (
	ErrRealNameAuthSuccess    = 0 //成功
	ErrRealNameAuthFail       = 1 //失败
	ErrRealNameAuthAlready    = 2 //已经实名认证
	ErrRealNameAuthBusiness   = 3 //三方接口未通过
	ErrRealNameAuthLengthLack = 4 //身份证号码长度不足
	ErrRealNameAuthLengthMore = 5 //身份证号码长度过长
)

type S2C_RealNameAuth struct {
	RealName string
	Error    int
	ErrMsg   string
}

type C2S_BindBankCard struct {
	BankName    string
	BankCardNo  string
	Province    string
	City        string
	OpeningBank string
}

const (
	ErrAddBankCardSuccess  = 0 //成功
	ErrAddBankCardFail     = 1 //失败
	ErrAddBankCardAlready  = 2 //已经绑定
	ErrAddBankCardBusiness = 3 //三方接口未通过
)

type S2C_BindBankCard struct {
	BankCardInfo *BankCardInfo
	Error        int
	ErrMsg       string
}

type BankCardInfo struct {
	BankName       string //银行名称
	BankCardNoTail string //银行卡号后四位
}

const (
	FlowTypeAward    = 1
	FlowTypeWithDraw = 2
	FlowTypeGift     = 3
)

type WithDrawData struct {
	FlowType  int
	MatchID   string
	Amount    float64
	Status    string
	CreatedAt int64
}

type S2C_AwardInfo struct {
	Amount       float64
	WithDrawList []WithDrawData
}

type C2S_WithDraw struct {
}

const (
	ErrWithDrawSuccess = 0
	ErrWithDrawFail    = 1
	ErrWithDrawLack    = 2
	ErrWithDrawNoAuth  = 3
	ErrWithDrawNoBank  = 4
	ErrWithDrawMore    = 5
)

type S2C_WithDraw struct {
	Amount float64
	Error  int
	ErrMsg string
}

type S2C_OnlineUserNum struct {
	Num int
}
