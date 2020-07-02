package session

import (
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	. "ddz/game/room"
	"ddz/msg"
	"reflect"

	"github.com/name5566/leaf/gate"
	"github.com/szxby/tools/log"
)

func init() {

	handler(&msg.C2S_Heartbeat{}, handleHeartbeat)

	handler(&msg.C2S_GetAllPlayers{}, handleGetAllPlayers)

	handler(&msg.C2S_LandlordBid{}, handleLandlordBidScore)

	handler(&msg.C2S_LandlordDouble{}, handleLandlordDouble)

	handler(&msg.C2S_LandlordDiscard{}, handleLandlordDiscard)

	handler(&msg.C2S_LandlordMatchRound{}, handleGetRank)

	handler(&msg.C2S_GetGameRecord{}, handleGetGameRecord)
	handler(&msg.C2S_GetGameRankRecord{}, handleGetGameRankRecord)
	handler(&msg.C2S_GetGameResultRecord{}, handleGetGameResultRecord)

	handler(&msg.C2S_SetNickName{}, handleNickName)

	handler(&msg.C2S_SystemHost{}, handleSystemHost)
	handler(&msg.C2S_Apply{}, handleApply)
	handler(&msg.C2S_GetCoupon{}, handleCoupon)
	handler(&msg.C2S_DailySign{}, handleDailySign)
	handler(&msg.C2S_RaceDetail{}, handleRaceDetail)
	handler(&msg.C2S_FeedBack{}, handleFeedBack)
	handler(&msg.C2S_ReadMail{}, handleReadMail)
	handler(&msg.C2S_DeleteMail{}, handleDeleteMail)
	handler(&msg.C2S_TakenMailAnnex{}, handleTakenMailAnnex)

	handler(&msg.C2S_RankingList{}, handleRankingList)
	handler(&msg.C2S_RealNameAuth{}, handleRealNameAuth)
	handler(&msg.C2S_BindBankCard{}, handleAddBankCard)
	handler(&msg.C2S_AwardInfo{}, handleAwardInfo)
	handler(&msg.C2S_WithDraw{}, handleWithDraw)
	handler(&msg.C2S_GetMatchList{}, handleGetMatchList)
	handler(&msg.C2S_ChangePassword{}, handleChangePassword)
	handler(&msg.Test_WriteFlowData{}, handleTest)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleHeartbeat(args []interface{}) {
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.User == nil {
		return
	}
	agentInfo.User.HeartbeatStop = false
}
func handleCoupon(args []interface{}) {
	m := args[0].(*msg.C2S_GetCoupon)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.User == nil {
		return
	}
	user := agentInfo.User
	hall.AddCoupon(user, m.Count)
}

func handleGetGameRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetGameRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if m.PageNumber < 1 {
		m.PageNumber = 1
	}
	if m.PageSize < 1 {
		m.PageSize = 10
	}
	user.SendMatchRecord(m.PageNumber, m.PageSize)
}

func handleGetGameRankRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetGameRankRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if m.PageNumber < 1 {
		m.PageNumber = 1
	}
	if m.PageSize < 1 {
		m.PageSize = 10
	}
	user.SendMatchRankRecord(m.MatchID, m.PageNumber, m.PageSize, m.RankNumber, m.RankSize)
}

func handleGetGameResultRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetGameResultRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if m.PageNumber < 1 {
		m.PageNumber = 1
	}
	if m.PageSize < 1 {
		m.PageSize = 10
	}
	user.SendMatchResultRecord(m.MatchID, m.PageNumber, m.PageSize, m.ResultNumber, m.ResultSize)
}

func handleNickName(args []interface{}) {
	m := args[0].(*msg.C2S_SetNickName)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.SetNickname(user, m.NickName)
}

func handleGetAllPlayers(args []interface{}) {
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.User == nil {
		return
	}
	user := agentInfo.User
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		r.Play(args[0], user.BaseData.UserData.UserID)
	}
}

func handleLandlordBidScore(args []interface{}) {
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		r.Play(args[0], user.BaseData.UserData.UserID)
	}

}
func handleLandlordDouble(args []interface{}) {
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		r.Play(args[0], user.BaseData.UserData.UserID)
	}
}
func handleGetRank(args []interface{}) {
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		a.Close()
		return
	}
	// if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
	// 	r.Play(args[0], user.BaseData.UserData.UserID)
	// }
	uid := user.BaseData.UserData.UserID
	match, ok := UserIDMatch[uid]
	if !ok {
		log.Error("player %v not in match", uid)
		return
	}
	match.GetRank(uid)
}

func handleLandlordDiscard(args []interface{}) {
	_ = args[0].(*msg.C2S_LandlordDiscard)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		r.Play(args[0], user.BaseData.UserData.UserID)
	}
}

func handleSystemHost(args []interface{}) {
	_ = args[0].(*msg.C2S_SystemHost)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {

		r.Play(args[0], user.BaseData.UserData.UserID)
	}
}

func handleDailySign(args []interface{}) {
	_ = args[0].(*msg.C2S_DailySign)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		a.Close()
		return
	}
	hall.DailySign(user)
}

func handleRaceDetail(args []interface{}) {
	if len(args) != 2 {
		return
	}

	m := args[0].(*msg.C2S_RaceDetail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	if m, ok := MatchManagerList[m.ID]; ok {
		m.SendMatchDetail(user.BaseData.UserData.UserID)
		// if s, ok := UserIDMatch[user.BaseData.UserData.UserID]; ok {
		// 	if m.ID == s.MatchID {
		// 		data.IsSign = true
		// 	}
		// }

		// user.WriteMsg(data)
	}
}

func handleFeedBack(args []interface{}) {
	if len(args) != 2 {
		return
	}

	m := args[0].(*msg.C2S_FeedBack)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.Feedback(user, m.Title, m.Content)
}

func handleReadMail(args []interface{}) {
	if len(args) != 2 {
		return
	}
	m := args[0].(*msg.C2S_ReadMail)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.ReadMail(m.ID)
	hall.SendMail(user)
}

func handleDeleteMail(args []interface{}) {
	if len(args) != 2 {
		return
	}
	m := args[0].(*msg.C2S_DeleteMail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.DeleteMail(m.ID)
	user.WriteMsg(&msg.S2C_DeleteMail{
		Error: msg.S2C_DeleteMail_OK,
	})
	hall.SendMail(user)
}

func handleTakenMailAnnex(args []interface{}) {
	if len(args) != 2 {
		return
	}
	m := args[0].(*msg.C2S_TakenMailAnnex)
	a := args[1].(gate.Agent)

	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.TakenMailAnnex(m.ID)
	user.WriteMsg(&msg.S2C_DeleteMail{
		Error: msg.S2C_TakenMailAnnex_OK,
	})
	hall.SendMail(user)
}

func handleApply(args []interface{}) {
	m := args[0].(*msg.C2S_Apply)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		a.Close()
		return
	}
	//验证赛事ID的合法性
	v, ok := MatchManagerList[m.MatchId]
	if !ok {
		return
	}
	if m.Action == 1 {
		v.SignIn(user.BaseData.UserData.UserID)
	} else if m.Action == 2 {
		v.SignOut(user.BaseData.UserData.UserID, m.MatchId)
	}
}

func handleRankingList(args []interface{}) {
	m := args[0].(*msg.C2S_RankingList)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.SendRankingList(user)
}

func handleRealNameAuth(args []interface{}) {
	m := args[0].(*msg.C2S_RealNameAuth)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.RealNameAuth(user, m)
}

func handleAddBankCard(args []interface{}) {
	m := args[0].(*msg.C2S_BindBankCard)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.AddBankCard(user, m)
}

func handleAwardInfo(args []interface{}) {
	m := args[0].(*msg.C2S_AwardInfo)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.SendAwardInfo(user)
}

func handleWithDraw(args []interface{}) {
	m := args[0].(*msg.C2S_WithDraw)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.WithDraw(user)
}

func handleGetMatchList(args []interface{}) {
	// m := args[0].(*msg.C2S_GetMatchList)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	myMatchID := ""
	if ma, ok := UserIDMatch[user.BaseData.UserData.UserID]; ok {
		// myMatchID = ma.MatchID
		myMatchID = ma.NormalCofig.MatchID
	}
	list := GetMatchManagerInfo(2).([]msg.OneMatch)
	for i, v := range list {
		// 已报名的比赛排序在最前面
		if v.MatchID == myMatchID {
			list[i].IsSign = true
			list[i], list[0] = list[0], list[i]
			break
		}
	}
	user.WriteMsg(&msg.S2C_GetMatchList{
		List: list,
	})
}

func handleChangePassword(args []interface{}) {
	m := args[0].(*msg.C2S_ChangePassword)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	hall.ChangePassword(user, m)
}

func handleTest(args []interface{}) {
	log.Debug("【写入成功】")
	m := args[0].(*msg.Test_WriteFlowData)

	uid := m.UID
	ud := ReadUserDataByID(uid)
	ud.Fee += 10
	SaveUserData(ud)
	hall.WriteFlowData(uid, 10, hall.FlowTypeAward, "AAAATest","xxxx!!!", []int{})
}