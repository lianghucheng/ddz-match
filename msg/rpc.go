package msg

func init() {
	Processor.Register(&C2S_EndMatch{})
	Processor.Register(&RPC_SendMail{})
	Processor.Register(&RPC_SendRaceInfo{})
	Processor.Register(&RPC_WriteAwardFlowData{})
	Processor.Register(&RPC_SendMatchEndMail{})
	Processor.Register(&RPC_SendInterruptMail{})
}

type C2S_EndMatch struct {
	MatchId string //赛事ID
	Id      int    //用户ID
}

type RPC_SendMail struct {
	ID int //用户ID
}

type RPC_SendMatchEndMail struct {
	Userid int
	Matchid string
	Order int
	Award float64
}

type RPC_SendInterruptMail struct {
	Userid int
	Matchid string
}

type RPC_SendRaceInfo struct {
	ID int //Userid
}

type RPC_WriteAwardFlowData struct {
	Userid  int
	Amount  float64
	Matchid string
}
