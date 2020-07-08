package msg

func init() {
	Processor.Register(&C2S_EndMatch{})
	Processor.Register(&RPC_SendMail{})
	Processor.Register(&RPC_SendRaceInfo{})
	Processor.Register(&RPC_WriteAwardFlowData{})
	Processor.Register(&RPC_SendMatchEndMail{})
	Processor.Register(&RPC_SendInterruptMail{})
	Processor.Register(&RPC_TempPayOK{})
	Processor.Register(&RPC_AddFee{})
	Processor.Register(&RPC_TestAddAward{})
	Processor.Register(&RPC_UpdateAwardInfo{})
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

type RPC_TempPayOK struct {
	TotalFee  int
	AccountID int
}

type RPC_AddFee struct {
	FeeType string  `json:"fee_type"`
	Userid  int     `json:"userid"`
	Amount  float64 `json:"amount"`
}

type RPC_TestAddAward struct {
	Uid 	int
	Amount  float64
}

type RPC_UpdateAwardInfo struct {
	Uid 	int
}