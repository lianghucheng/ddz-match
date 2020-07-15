package session

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	"ddz/msg"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"

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
	skeleton.RegisterChanRPC("AddFee", rpcAddFee)
	skeleton.RegisterChanRPC("AddAward", rpcAddAward)
	skeleton.RegisterChanRPC("UpdateAwardInfo", rpcUpdateAwardInfo)
	skeleton.RegisterChanRPC("optUser", optUser)     // 操作玩家
	skeleton.RegisterChanRPC("clearInfo", clearInfo) // 清除玩家实名信息
	skeleton.RegisterChanRPC("UpdateCoupon", rpcUpdateCoupon)
	skeleton.RegisterChanRPC("UpdateHeadImg", rpcUpdateHeadImg)
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
	_ = m
	//hall.WriteFlowData(m.Userid, m.Amount, hall.FlowTypeAward, MatchList[m.Matchid].NormalCofig.MatchType,[]int{})
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

	addCoupon := m.TotalFee / 100

	if user, ok := UserIDUsers[ud.UserID]; ok {
		user.GetUserData().Coupon += int64(addCoupon)
		go func() {
			SaveUserData(user.GetUserData())
		}()
		user.WriteMsg(&msg.S2C_GetCoupon{
			Error: msg.ErrPaySuccess,
		})
		hall.UpdateUserCoupon(user, int64(addCoupon), user.GetUserData().Coupon-int64(addCoupon), user.GetUserData().Coupon, db.ChargeOpt, db.Charge)
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

func rpcAddFee(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_AddFee)
	if user, ok := UserIDUsers[m.Userid]; ok {
		ud := user.GetUserData()
		if m.FeeType == "fee" {
			ud.Fee += m.Amount
		} else if m.FeeType == "takenfee" {
			ud.TakenFee += m.Amount
		}
		SaveUserData(ud)
		hall.UpdateUserAfterTaxAward(user)
		hall.SendAwardInfo(user)
	} else {
		ud := ReadUserDataByID(m.Userid)
		if m.FeeType == "fee" {
			ud.Fee += m.Amount
		} else if m.FeeType == "takenfee" {
			ud.TakenFee += m.Amount
		}
		SaveUserData(ud)
	}
}

func rpcAddAward(args []interface{}) {
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_AddAward)
	if user, ok := UserIDUsers[m.Uid]; ok {
		user.GetUserData().Fee += m.Amount
		game.GetSkeleton().Go(func() {
			SaveUserData(user.GetUserData())
			hall.WriteFlowData(m.Uid, m.Amount, hall.FlowTypeGift, "", "", []int{})
		}, func() {
			hall.UpdateUserAfterTaxAward(user)
		})
	} else {
		ud := ReadUserDataByID(m.Uid)
		ud.Fee += m.Amount
		game.GetSkeleton().Go(func() {
			SaveUserData(ud)
			hall.WriteFlowData(m.Uid, m.Amount, hall.FlowTypeGift, "", "", []int{})
		}, nil)
	}
	log.Debug("【添加提现测试数据成功】")
}

func rpcUpdateAwardInfo(args []interface{}) {
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_UpdateAwardInfo)
	if user, ok := UserIDUsers[m.Uid]; ok {
		hall.SendAwardInfo(user)
	}
}

func optUser(args []interface{}) {
	// log.Debug("optUser:%+v", args)
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_OptUser)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("optUser:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	if data.Opt != -1 && data.Opt != 1 {
		code = 1
		desc = "无效操作!"
		return
	}
	if user, ok := UserIDUsers[data.UID]; ok && data.Opt == -1 {
		user.Close()
	}
	skeleton.Go(func() {
		UpdateUserData(data.UID, bson.M{"$set": bson.M{"role": data.Opt}})
	}, nil)
}

func clearInfo(args []interface{}) {
	// log.Debug("optUser:%+v", args)
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_ClearInfo)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("data:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	if data.Opt != 1 && data.Opt != 2 { // 1清理实名信息＋银行卡信息，２只清理银行卡
		code = 1
		desc = "无效操作!"
		return
	}
	if user, ok := UserIDUsers[data.UID]; ok && data.Opt == 1 {
		user.BaseData.UserData.RealName = ""
		user.BaseData.UserData.IDCardNo = ""
	}
	bank := &hall.BankCard{
		Userid: data.UID,
	}
	skeleton.Go(func() {
		if data.Opt == 1 {
			UpdateUserData(data.UID, bson.M{"$set": bson.M{"realname": "", "idcardno": ""}})
		}
		db.UpdateBankInfo(data.UID, bank)
	}, nil)
}

func rpcUpdateCoupon(args []interface{}) {
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_UpdateCoupon)
	ud := ReadUserDataByAid(m.Accountid)
	if user, ok := UserIDUsers[ud.UserID]; ok {
		ud := user.GetUserData()
		ud.Coupon += int64(m.Amount)
		SaveUserData(ud)
		hall.UpdateUserCoupon(user, int64(m.Amount), ud.Coupon-int64(m.Amount), ud.Coupon, db.NormalOpt, db.Backstage)
	} else {
		ud.Coupon += int64(m.Amount)
		SaveUserData(ud)
	}
}

func rpcUpdateHeadImg(args []interface{}) {
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_UpdateHeadImg)
	ud := ReadUserDataByAid(m.Accountid)
	if user, ok := UserIDUsers[ud.UserID]; ok {
		ud := user.GetUserData()
		ud.Headimgurl = m.HeadImg
		SaveUserData(ud)
	} else {
		ud.Headimgurl = m.HeadImg
		SaveUserData(ud)
	}
}
