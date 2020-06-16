package session

import (
	"ddz/conf"
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	. "ddz/game/room"
	. "ddz/game/values"
	"ddz/msg"
	"strings"
	"time"

	"github.com/name5566/leaf/gate"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC("TokenLogin", rpcTokenLogin)
	skeleton.RegisterChanRPC("UsernamePasswordLogin", rpcUsernamePasswordLogin)
	skeleton.RegisterChanRPC("EndMatch", rpcEndMatch)
	skeleton.RegisterChanRPC("SendMail", rpcSendMail)
	skeleton.RegisterChanRPC("SendRaceInfo", rpcSendRaceInfo)
	skeleton.RegisterChanRPC("WriteAwardFlowData", rpcWriteAwardFlowData)
	skeleton.RegisterChanRPC("SendMatchEndMail", rpcSendMatchEndMail)
	skeleton.RegisterChanRPC("SendInterruptMail", rpcSendInterruptMail)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	a.SetUserData(new(AgentInfo))
	skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().ConnectTimeout)*time.Second, func() {
		if a.UserData() != nil {
			agentInfo := a.UserData().(*AgentInfo)
			if agentInfo != nil && agentInfo.User == nil {
				a.Close()
			}
		}
	})

}

func rpcTokenLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_TokenLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.User != nil {
		return
	}
	if strings.TrimSpace(m.Token) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_TokenInvalid})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).User = newUser
	tokenLogin(newUser, m.Token)
}

func rpcUsernamePasswordLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_AccountLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.User != nil {
		return
	}
	if strings.TrimSpace(m.Account) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UsernameInvalid})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).User = newUser
	usernamePasswordLogin(newUser, m.Account, m.Code)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	user := a.UserData().(*AgentInfo).User
	a.SetUserData(nil)
	if user == nil {
		return
	}
	if user.State == UserLogin {
		user.State = UserLogout
		logout(user)
	}
}

func rpcEndMatch(args []interface{}) {
	m := args[0].(*msg.C2S_EndMatch)
	delete(UserIDRooms, m.Id)
	if game, ok := UserIDRooms[m.Id]; ok {
		game.Exit(m.Id)
	}
	delete(UserIDMatch, m.Id)
	delete(MatchList[m.MatchId].AllPlayers, m.Id)
	Broadcast(&msg.S2C_MatchNum{
		MatchId: m.MatchId,
		Count:   len(MatchList[m.MatchId].AllPlayers),
	})
}

func rpcSendMail(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendMail)
	if m.ID == -1 {
		for _, v := range UserIDUsers {
			hall.SendMail(v)
		}
	} else {
		if user, ok := UserIDUsers[m.ID]; ok {
			hall.SendMail(user)
		}
	}
}

func rpcSendRaceInfo(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendRaceInfo)
	RaceInfo := []msg.RaceInfo{}
	for _, v := range MatchList {
		var award float64
		if len(v.Award) > 0 {
			award = v.Award[0]
		}
		RaceInfo = append(RaceInfo, msg.RaceInfo{
			ID:       v.MatchID,
			Desc:     v.MatchName,
			Award:    award,
			EnterFee: float64(v.EnterFee) / 100,
			ConDes:   v.MatchDesc,
			JoinNum:  len(MatchList[v.MatchID].SignInPlayers),
		})
	}
	if user, ok := UserIDUsers[m.ID]; ok {
		user.WriteMsg(&msg.S2C_RaceInfo{
			Races: RaceInfo,
		})
	}
}

func rpcWriteAwardFlowData(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_WriteAwardFlowData)

	hall.WriteFlowData(m.Userid, m.Amount, hall.FlowTypeAward, MatchList[m.Matchid].MatchType)
}

func rpcSendMatchEndMail(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendMatchEndMail)

	hall.MatchEndPushMail(m.Userid, MatchList[m.Matchid].MatchName, m.Order, m.Award)
}

func rpcSendInterruptMail(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendInterruptMail)

	hall.MatchInterruptPushMail(m.Userid, MatchList[m.Matchid].MatchName, int(MatchList[m.Matchid].EnterFee))
}
