package session

import (
	"ddz/conf"
	"ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	"ddz/msg"
	"fmt"
	"strings"
	"time"

	"github.com/szxby/tools/log"

	"github.com/name5566/leaf/gate"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC("TokenLogin", rpcTokenLogin)
	skeleton.RegisterChanRPC("UsernamePasswordLogin", rpcUsernamePasswordLogin)
	skeleton.RegisterChanRPC("AccountLogin", rpcAccountLogin)
	skeleton.RegisterChanRPC("SendMail", rpcSendMail)
	skeleton.RegisterChanRPC("SendRaceInfo", rpcSendRaceInfo)
	skeleton.RegisterChanRPC("WriteAwardFlowData", rpcWriteAwardFlowData)
	// skeleton.RegisterChanRPC("SendMatchEndMail", rpcSendMatchEndMail)
	skeleton.RegisterChanRPC("SendInterruptMail", rpcSendInterruptMail)
	skeleton.RegisterChanRPC("TempPayOK", rpcTempPayOK)
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
	m := args[1].(*msg.C2S_UsrnPwdLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.User != nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UsernameInvalid})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).User = newUser
	usernamePasswordLogin(newUser, m.Username, m.Password)
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
	RaceInfo := GetMatchManagerInfo(1).([]msg.RaceInfo)
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

	hall.WriteFlowData(m.Userid, m.Amount, hall.FlowTypeAward, MatchList[m.Matchid].NormalCofig.MatchType,[]int{})
}

// func rpcSendMatchEndMail(args []interface{}) {
// 	if len(args) != 1 {
// 		return
// 	}
// 	m := args[0].(*msg.RPC_SendMatchEndMail)
// 	hall.MatchEndPushMail(m.Userid, m.MatchName, m.Order, m.Award)
// }

func rpcSendInterruptMail(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendInterruptMail)

	hall.MatchInterruptPushMail(m.Userid, m.MatchName, m.Coupon)
}

func rpcTempPayOK(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_TempPayOK)

	fmt.Println("【！！！！！！！！！】Accountid:", m.AccountID)
	ud := ReadUserDataByAid(m.AccountID)

	if ud == nil {
		log.Error("不存在该用户")
		return
	}

	addCoupon := m.TotalFee / 10

	if user, ok := UserIDUsers[ud.UserID]; ok {
		user.GetUserData().Coupon += int64(addCoupon)
		go func() {
			SaveUserData(user.GetUserData())
		}()
		user.WriteMsg(&msg.S2C_GetCoupon{
			Error: msg.ErrPaySuccess,
		})
		hall.UpdateUserCoupon(user, int64(addCoupon), db.Charge)
	} else {
		ud.Coupon += int64(addCoupon)
		go func() {
			SaveUserData(ud)
		}()
	}
}

func rpcAccountLogin(args []interface{}) {
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
	AccountLogin(newUser, m.Account, m.Code)
}
