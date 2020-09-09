package session

import (
	"ddz/conf"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	"ddz/game/server"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"fmt"
	"strconv"
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
	skeleton.RegisterChanRPC("NotifyPayOK", rpcNotifyPayOK)
	skeleton.RegisterChanRPC("AddFee", rpcAddFee)
	skeleton.RegisterChanRPC("AddAward", rpcAddAward)
	skeleton.RegisterChanRPC("UpdateAwardInfo", rpcUpdateAwardInfo)
	skeleton.RegisterChanRPC("optUser", optUser)             // 操作玩家
	skeleton.RegisterChanRPC("clearInfo", clearInfo)         // 清除玩家实名信息
	skeleton.RegisterChanRPC("restartServer", restartServer) // 服务器停服
	skeleton.RegisterChanRPC("editWhiteList", editWhiteList) // 白名单操作
	skeleton.RegisterChanRPC("getOnline", getOnline)         // 获取在线人数
	skeleton.RegisterChanRPC("UpdateCoupon", rpcUpdateCoupon)
	skeleton.RegisterChanRPC("UpdateHeadImg", rpcUpdateHeadImg)
	skeleton.RegisterChanRPC("AddCouponFrag", rpcAddCouponFrag)
	skeleton.RegisterChanRPC("SendPayAccount", rpcSendPayAccount)
	skeleton.RegisterChanRPC("UpdateBankCardNo", rpcUpdateBankCardNo)
	skeleton.RegisterChanRPC("dealIllegalMatch", dealIllegalMatch)
	skeleton.RegisterChanRPC("shareAward", shareAward)
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
	RaceInfo := GetMatchManagerInfo(2).([]msg.OneMatch)
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

func rpcNotifyPayOK(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_NotifyPayOK)

	ud := ReadUserDataByAid(m.AccountID)

	if ud == nil {
		log.Error("不存在该用户")
		return
	}

	// 充值返利
	game.GetSkeleton().Go(func() {
		if err := utils.PostToAgentServer(struct {
			AccountID int
			Amount    int64
		}{
			AccountID: m.AccountID,
			Amount:    int64(m.TotalFee),
		}, "/rebate"); err != nil {
			log.Error("rebate err:%v", err)
		}
	}, nil)

	//发货
	addCoupon := m.TotalFee / 100
	if m.Amount > 0 {
		addCoupon = m.Amount
	}

	hall.SendGoods(ud.UserID, addCoupon)
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
	ud := ReadUserDataByID(m.Userid)
	if user, ok := UserIDUsers[m.Userid]; ok {
		ud := user.GetUserData()
		if m.FeeType == "fee" {
			ud.Fee = hall.FeeAmount(ud.UserID)
		} else if m.FeeType == "takenfee" {
			ud.TakenFee += m.Amount
		}
		SaveUserData(ud)
		hall.UpdateUserAfterTaxAward(user)
		hall.SendAwardInfo(user)
	} else {
		if m.FeeType == "fee" {
			ud.Fee += hall.FeeAmount(ud.UserID)
		} else if m.FeeType == "takenfee" {
			ud.TakenFee += m.Amount
		}
		SaveUserData(ud)
	}
	bankCard := new(hall.BankCard)
	bankCard.Userid = ud.UserID
	bankCard.Read()
	if m.FeeType == "fee" {
		hall.RefundPushMail(ud.UserID, m.Amount)
	} else if m.FeeType == "takenfee" {
		hall.PrizePresentationPushMail(ud.UserID, bankCard.BankName, m.Amount)
	}
}

func rpcAddAward(args []interface{}) {
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_AddAward)
	ud := ReadUserDataByAid(m.Uid)
	if ud.UserID == 0 {
		log.Error("unknown user:%+v", m)
		return
	}
	hall.WriteFlowData(m.Uid, m.Amount, hall.FlowTypeGift, "", "", []int{}, nil)
	hall.AddFee(ud.UserID, ud.AccountID, m.Amount, db.PlatformOpt, db.Backstage, "")

	// if user, ok := UserIDUsers[m.Uid]; ok {
	// 	hall.WriteFlowData(m.Uid, m.Amount, hall.FlowTypeGift, "", "", []int{})
	// 	user.GetUserData().Fee = hall.FeeAmount(m.Uid)
	// 	game.GetSkeleton().Go(func() {
	// 		SaveUserData(user.GetUserData())
	// 	}, func() {
	// 		hall.UpdateUserAfterTaxAward(user)
	// 	})
	// } else {
	// 	hall.WriteFlowData(m.Uid, m.Amount, hall.FlowTypeGift, "", "", []int{})
	// 	ud := ReadUserDataByID(m.Uid)
	// 	ud.Fee = hall.FeeAmount(m.Uid)
	// 	game.GetSkeleton().Go(func() {
	// 		SaveUserData(ud)
	// 	}, func() {

	// 	})
	// }
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
		user.BaseData.UserData.Role = RoleBlack
		user.Close()
	}
	// skeleton.Go(func() {
	// 	UpdateUserData(data.UID, bson.M{"$set": bson.M{"role": data.Opt}})
	// }, nil)
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
	if ud.UserID == 0 {
		log.Error("unknown user:%+v", m)
		return
	}
	// if user, ok := UserIDUsers[ud.UserID]; ok {
	// 	ud := user.GetUserData()
	// 	ud.Coupon += int64(m.Amount)
	// 	SaveUserData(ud)
	// 	hall.UpdateUserCoupon(user, int64(m.Amount), ud.Coupon-int64(m.Amount), ud.Coupon, db.NormalOpt, db.Backstage)
	// } else {
	// 	ud.Coupon += int64(m.Amount)
	// 	SaveUserData(ud)
	// }
	hall.AddCoupon(ud.UserID, ud.AccountID, int64(m.Amount), db.PlatformOpt, db.Backstage, "")
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

func rpcAddCouponFrag(args []interface{}) {
	log.Debug("远程调用加点券碎片")
	if len(args) != 1 {
		log.Debug("参数长度异常")
		return
	}
	m := args[0].(*msg.RPC_AddCouponFrag)
	if m.Secret != "123456" {
		log.Debug("非法请求")
		return
	}
	ud := ReadUserDataByAid(m.Accountid)
	if ud.UserID == 0 {
		log.Error("unknown user:%+v", m)
		return
	}
	// hall.AddPropAmount(config.PropTypeCouponFrag, m.Accountid, m.Amount)
	hall.AddFragment(ud.UserID, ud.AccountID, int64(m.Amount), db.PlatformOpt, db.Backstage, "")
	log.Debug("成功！！！远程调用加点券碎片")
}

func restartServer(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_Restart)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("restart data:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	if values.DefaultRestartConfig.Status == values.RestartStatusIng && data.Status != values.RestartStatusFinish {
		code = 1
		desc = "服务器维护进行中,无法修改!"
		return
	}
	// if values.DefaultRestartConfig.Status >= values.RestartStatusFinish {
	// 	code = 1
	// 	desc = "服务器维护已完成,无法修改!"
	// 	return
	// }
	if data.Status == values.RestartStatusFinish {
		values.DefaultRestartConfig.Status = values.RestartStatusFinish
		return
	}
	if err := db.GetRestart(); err != nil {
		code = 1
		desc = err.Error()
	}
	if values.DefaultRestartConfig.RestartTimer != nil {
		values.DefaultRestartConfig.RestartTimer.Stop()
	}
	if values.DefaultRestartConfig.Status < values.RestartStatusIng {
		startKick := 1 * time.Second
		if values.DefaultRestartConfig.RestartTime > time.Now().Unix() {
			startKick = time.Unix(values.DefaultRestartConfig.RestartTime, 0).Sub(time.Now())
		}
		values.DefaultRestartConfig.RestartTimer = game.GetSkeleton().AfterFunc(startKick, func() {
			values.DefaultRestartConfig.Status = values.RestartStatusIng
			server.KickAllPlayers()
			db.UpdateRestart(bson.M{"id": values.DefaultRestartConfig.ID}, bson.M{"$set": bson.M{"status": values.RestartStatusIng}})
			// 自动打开白名单
			values.DefaultWhiteListConfig.WhiteSwitch = true
			db.UpdateWhite(true)
			log.Debug("white:%+v,restart:%+v", values.DefaultWhiteListConfig, values.DefaultRestartConfig)
		})
	}
}

func editWhiteList(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_EditWhiteList)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("edit white data:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	if err := db.GetWhiteList(); err != nil {
		code = 1
		desc = "更新失败！"
		return
	}
	log.Debug("update white:%+v", values.DefaultWhiteListConfig)
}

func getOnline(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_GetOnline)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("online data:%+v", data)
	code := 0
	desc := "OK"
	onlinePlayer := 0
	matchPlayer := 0
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc, "online": onlinePlayer, "match": matchPlayer})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	onlinePlayer = len(UserIDUsers)
	matchPlayer = len(UserIDMatch)
}

func rpcSendPayAccount(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_SendPayAccount)
	_ = m
	hall.SendPayAccount(nil, hall.SendBroacast)
}

func rpcUpdateBankCardNo(args []interface{}) {
	if len(args) != 1 {
		return
	}
	m := args[0].(*msg.RPC_UpdateBankCardNo)
	user, ok := UserIDUsers[m.Userid]
	if ok {
		user.GetUserData().BankCardNo = m.BankCardNo
		SaveUserData(user.GetUserData())
	} else {
		ud := ReadUserDataByID(m.Userid)
		ud.BankCardNo = m.BankCardNo
		SaveUserData(ud)
	}
}

func shareAward(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_ShareAward)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("share award data:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	err := hall.ShareAwardPushMail(data.AccountID, data.Item, data.AwardNum)
	if err != nil {
		code = 1
		desc = err.Error()
	}
}

func dealIllegalMatch(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_IllegalMatch)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("illegalMatch data:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	matchID := []byte(data.SonMatchID)
	if len(matchID) < 20 {
		code = 1
		desc = "请求赛事ID有误!"
		return
	}
	round := []byte(fmt.Sprintf("%02d", data.Round))
	matchID[len(matchID)-8] = round[0]
	matchID[len(matchID)-7] = round[1]
	// 拉取体总发奖结果
	results, err := edy_api.AwardResult(string(matchID), 1, 100)
	if err != nil {
		code = 1
		desc = "回调体总失败！"
		return
	}
	var awardAmount float64
	for _, v := range results.Result_list {
		if v.Player_id == strconv.Itoa(data.AccountID) {
			if v.Status == "3" {
				log.Error("err award :%+v", v)
				code = 2
				desc = "体总已驳回!"
				return
			}
			if v.Status == "4" {
				log.Error("err award :%+v", v)
				code = 3
				desc = "体总返回异常赛事!"
				return
			}
			if v.Status != "2" {
				log.Error("err award :%+v", v)
				code = 1
				desc = "体总未发奖！"
				return
			}
			var err error
			awardAmount, err = strconv.ParseFloat(v.Bonous, 64)
			if err != nil {
				log.Error("err award :%+v", v)
				code = 1
				desc = "体总回调数据出错！"
				return
			}
			break
		}
	}
	if awardAmount <= 0 {
		code = 1
		desc = "回调体总失败！"
		return
	}

	fid := 0

	overFlowdata := db.ReadFlowdataLateOver(data.CreateTime)
	end := int64(0)
	if overFlowdata == nil {
		fid = hall.WriteFlowDataWithTime(data.UID, utils.Decimal(awardAmount), hall.FlowTypeAward,
			data.MatchType, data.SonMatchID, []int{}, data.CreateTime, hall.FlowDataStatusNormal)
		if fid <= 0 {
			code = 1
			desc = "回调体总失败！"
			return
		}
		end = values.MAX
	} else {
		fid = hall.WriteFlowDataWithTime(data.UID, utils.Decimal(awardAmount), hall.FlowTypeAward,
			data.MatchType, data.SonMatchID, []int{}, data.CreateTime, hall.FlowDataStatusOver)
		if fid <= 0 {
			code = 1
			desc = "回调体总失败！"
			return
		}
		overFlowdata.FlowIDs = append(overFlowdata.FlowIDs, fid)
		overFlowdata.ChangeAmount += utils.Decimal(awardAmount)
		db.SaveFlowdata(overFlowdata)

		end = overFlowdata.CreatedAt
	}

	backFlowdatas := db.ReadFlowdataBack(data.CreateTime, end)
	if backFlowdatas != nil {
		for _, v := range *backFlowdatas {
			v.FlowIDs = append(v.FlowIDs, fid)
			v.ChangeAmount += utils.Decimal(awardAmount)
			db.SaveFlowdata(&v)
		}
	}
	hall.AddFeeWithTime(data.UID, data.AccountID, utils.Decimal(awardAmount),
		db.MatchOpt, db.MatchAward+fmt.Sprintf("-%v", data.MatchName), data.SonMatchID, data.CreateTime)
	//hall.WriteMatchAwardRecordWithTime(data.UID, data.MatchType, data.MatchID, data.MatchName, data.Award, data.CreateTime)
	db.UpdateIllegalMatchRecord(bson.M{"accountid": data.AccountID, "sonmatchid": data.SonMatchID}, bson.M{"$set": bson.M{"callbackstatus": 3}})
}
