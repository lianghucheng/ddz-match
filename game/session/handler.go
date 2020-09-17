package session

import (
	"ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/match"
	"ddz/game/pay"
	. "ddz/game/player"
	. "ddz/game/room"
	"ddz/game/values"
	"ddz/msg"
	"github.com/name5566/leaf/gate"
	"github.com/szxby/tools/log"
	"reflect"
)

func init() {

	handler(&msg.C2S_Heartbeat{}, handleHeartbeat)

	handler(&msg.C2S_GetAllPlayers{}, handleGetAllPlayers)

	handler(&msg.C2S_LandlordBid{}, handleLandlordBidScore)

	handler(&msg.C2S_LandlordDouble{}, handleLandlordDouble)

	handler(&msg.C2S_LandlordDiscard{}, handleLandlordDiscard)

	handler(&msg.C2S_LandlordMatchRound{}, handleGetRank)

	handler(&msg.C2S_GetGameRecordAll{}, handleGetGameRecordAll)
	handler(&msg.C2S_GetGameRecord{}, handleGetGameRecord)
	handler(&msg.C2S_GetGameRankRecord{}, handleGetGameRankRecord)
	handler(&msg.C2S_GetGameResultRecord{}, handleGetGameResultRecord)

	handler(&msg.C2S_SetNickName{}, handleNickName)

	handler(&msg.C2S_SystemHost{}, handleSystemHost)
	handler(&msg.C2S_Apply{}, handleApply)
	handler(&msg.C2S_GetCoupon{}, handleCoupon)
	handler(&msg.C2S_DailySign{}, handleDailySign)
	handler(&msg.C2S_RaceInfo{}, handleRaceInfoHall)
	handler(&msg.C2S_RaceDetail{}, handleRaceDetail)
	handler(&msg.C2S_FeedBack{}, handleFeedBack)
	handler(&msg.C2S_ReadMail{}, handleReadMail)
	handler(&msg.C2S_DeleteMail{}, handleDeleteMail)
	handler(&msg.C2S_TakenMailAnnex{}, handleTakenMailAnnex)

	handler(&msg.C2S_RankingList{}, handleRankingList)
	handler(&msg.C2S_RealNameAuth{}, handleRealNameAuth)
	handler(&msg.C2S_BindBankCard{}, handleAddBankCard)
	handler(&msg.C2S_WithDraw{}, handleWithDraw)
	handler(&msg.C2S_GetMatchList{}, handleGetMatchList)
	handler(&msg.C2S_GetMatchSignList{}, handleGetMatchSignList)
	handler(&msg.C2S_ChangePassword{}, handleChangePassword)
	handler(&msg.C2S_TakenFirstCoupon{}, handleTakenFirstCoupon)
	handler(&msg.C2S_CreateEdyOrder{}, handleCreateEdyOrder)
	handler(&msg.C2S_CreateOrderSuccess{}, handleCreateOrderSuccess)
	handler(&msg.C2S_UseProp{}, handleUseProp)
	handler(&msg.C2S_Knapsack{}, handleKnapsack)
	handler(&msg.C2S_UserInfo{}, handleGetUserInfo)
	handler(&msg.C2S_GetDailyWelfareInfo{}, handleGetDailyWelfareInfo)
	handler(&msg.C2S_DrawDailyWelfareInfo{}, handleDrawDailyWelfareInfo)
	handler(&msg.C2S_TakenAllMail{}, handleTakenAllMail)
	handler(&msg.C2S_GetAllMail{}, handleGetSetMail)
	handler(&msg.C2S_DeleteAllMail{}, handleDeleteAllMail)
	handler(&msg.C2S_ActivityClick{}, handleActivityClick)
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
	hall.AddCoupon(user.BaseData.UserData.UserID, user.BaseData.UserData.AccountID, m.Count, db.NormalOpt, db.FakeCharge, "")
}

func handleGetGameRecordAll(args []interface{}) {
	// m := args[0].(*msg.C2S_GetGameRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	user.SendMatchRecordAll()
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
	user.SendMatchRecord(m.PageNumber, m.PageSize, m.MatchType)
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
	user.SendMatchResultRecord(m.MatchID, m.PageNumber, m.PageSize)
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
	} else {
		user.WriteMsg(&msg.S2C_RaceDetail{
			Error: 1,
		})
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
	user.WriteMsg(&msg.S2C_TakenMailAnnex{
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

func handleRaceInfoHall(args []interface{}) {
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	RaceInfo := GetMatchManagerInfo(2).([]msg.OneMatch)
	for i, v := range RaceInfo {
		if IsSign(user.BaseData.UserData.UserID, v.MatchID) {
			RaceInfo[i].IsSign = true
		}
	}
	user.WriteMsg(&msg.S2C_RaceInfo{
		Races: RaceInfo,
	})
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
	// myMatchID := ""
	// if ma, ok := UserIDMatch[user.BaseData.UserData.UserID]; ok {
	// 	// myMatchID = ma.MatchID
	// 	myMatchID = ma.NormalCofig.MatchID
	// }
	list := GetMatchManagerInfo(2).([]msg.OneMatch)
	sendConfig := []msg.OneMatchType{}
	sendList := []msg.OneMatch{}
	for i, v := range list {
		if c, ok := values.MatchTypeConfig[v.MatchType]; ok {
			tag := false
			for _, one := range sendConfig {
				if v.MatchType == one.MatchType {
					tag = true
					break
				}
			}
			if !tag {
				sendConfig = append(sendConfig, c)
			}
		}
		// // 已报名的比赛排序在最前面
		// if v.MatchID == myMatchID {
		// 	list[i].IsSign = true
		// 	list[i], list[0] = list[0], list[i]
		// }
		if IsSign(user.BaseData.UserData.UserID, v.MatchID) {
			list[i].IsSign = true
			sendList = append(sendList, list[i])
		}
	}
	for _, v := range list {
		if v.IsSign {
			continue
		}
		sendList = append(sendList, v)
	}

	user.WriteMsg(&msg.S2C_GetMatchList{
		All:  sendConfig,
		List: sendList,
	})
}

func handleGetMatchSignList(args []interface{}) {
	data := args[0].(*msg.C2S_GetMatchSignList)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	list := []SignList{}

	if v, ok := UserIDSign[user.BaseData.UserData.UserID]; ok {
		list = v
	}

	user.WriteMsg(&msg.S2C_GetMatchSignList{
		MatchID: data.MatchID,
		List:    list,
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

func handleTakenFirstCoupon(args []interface{}) {
	m := args[0].(*msg.C2S_TakenFirstCoupon)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.TakenFirstCoupon(user)
}

func handleCreateEdyOrder(args []interface{}) {
	m := args[0].(*msg.C2S_CreateEdyOrder)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}

	pay.CreateOrderTiZong(user, m)
}

func handleCreateOrderSuccess(args []interface{}) {
	m := args[0].(*msg.C2S_CreateOrderSuccess)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	pay.CreateOrderSuccess(user, m)
}

func handleUseProp(args []interface{}) {
	m := args[0].(*msg.C2S_UseProp)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.UseProp(user, m)
}

func handleKnapsack(args []interface{}) {
	m := args[0].(*msg.C2S_Knapsack)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	hall.SendKnapsack(user)
}

func handleGetUserInfo(args []interface{}) {
	// m := args[0].(*msg.C2S_UserInfo)
	// _ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	user.SendUserInfo()
}

func handleGetDailyWelfareInfo(args []interface{}) {
	// m := args[0].(*msg.C2S_GetDailyWelfareInfo)
	// _ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	user.GetDailyWelfareInfo()
}

func handleDrawDailyWelfareInfo(args []interface{}) {
	m := args[0].(*msg.C2S_DrawDailyWelfareInfo)
	// _ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	// user.DrawDailyWelfare(m.DailyType, m.AwardIndex)
	hall.DrawDailyWelfare(user, m.DailyType, m.AwardIndex)
}
func handleTakenAllMail(args []interface{}) {
	m := args[0].(*msg.C2S_TakenAllMail)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		log.Error("leaf UserData is nil. ")
		return
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		log.Error("system UserData is nil. ")
		return
	}

	hall.TakenAllMail(user)
}

func handleGetSetMail(args []interface{}) {
	m := args[0].(*msg.C2S_GetAllMail)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		log.Error("leaf UserData is nil. ")
		return
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		log.Error("system UserData is nil. ")
		return
	}

	hall.SendMail(user)
}

func checkAgent(a gate.Agent) (*User, bool) {
	if a.UserData() == nil {
		log.Error("leaf UserData is nil. ")
		return nil, false
	}

	user := a.UserData().(*AgentInfo).User
	if user == nil {
		log.Error("system UserData is nil. ")
		return nil, false
	}
	return user, true
}

func handleDeleteAllMail(args []interface{}) {
	m := args[0].(*msg.C2S_DeleteAllMail)
	_ = m
	a := args[1].(gate.Agent)
	if user, ok := checkAgent(a); !ok {
		return
	} else {
		hall.DeleteAllMail(user)
	}
}

func handleActivityClick(args []interface{}) {
	m := args[0].(*msg.C2S_ActivityClick)
	_ = m
	a := args[1].(gate.Agent)
	if _, ok := checkAgent(a); !ok {
		return
	} else {
		hall.AddActivityCnt(m.ID)
	}
}
