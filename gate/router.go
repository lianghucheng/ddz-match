package gate

import (
	"ddz/game"
	"ddz/login"
	"ddz/msg"
)

func init() {
	// login

	msg.Processor.SetRouter(&msg.C2S_TokenLogin{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AccountLogin{}, login.ChanRPC)
	// game

	msg.Processor.SetRouter(&msg.C2S_Heartbeat{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_GetAllPlayers{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_LandlordBid{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_LandlordDouble{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_LandlordDiscard{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_SystemHost{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_DailySign{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_Apply{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_RaceDetail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_FeedBack{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ReadMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_DeleteMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakenMailAnnex{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_DeleteMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakenMailAnnex{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordMatchRound{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetCoupon{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetGameRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetNickName{}, game.ChanRPC)
	//rpc

	msg.Processor.SetRouter(&msg.C2S_EndMatch{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.RPC_SendMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.RPC_SendRaceInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_RankingList{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_RealNameAuth{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AddBankCard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AwardInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_WithDraw{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.RPC_WriteAwardFlowData{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.RPC_SendMatchEndMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.RPC_SendInterruptMail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetMatchList{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetGameRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetGameRankRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetGameResultRecord{}, game.ChanRPC)
}