package hall

import (
	"ddz/conf"
	"ddz/config"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"fmt"
	"github.com/szxby/tools/log"
	"time"
)

func WithDraw(user *player.User) {
	checkWithdraw(user)
	if user.GetUserData().IsWithdraw {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: utils.Decimal(FeeAmount(user.UID())),
			Error:  msg.ErrWithDrawMore,
			ErrMsg: "每天只能提奖一次哦～",
		})
		return
	}
	withDraw(user, edy_api.WithDrawAPI)
}

func withDraw(user *player.User, callWithDraw func(userid int, amount float64) error) {
	fee := utils.Decimal(FeeAmount(user.UID()))
	if callWithDraw == nil {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: fee,
			Error:  msg.ErrWithDrawFail,
		})
		return
	}

	if user.RealName() == "" && !user.IsTest() {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: fee,
			Error:  msg.ErrWithDrawNoAuth,
			ErrMsg: "未实名认证",
		})
		return
	}

	if user.BankCardNo() == "" && !user.IsTest() {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: fee,
			Error:  msg.ErrWithDrawNoBank,
			ErrMsg: "未绑定银行卡",
		})
		return
	}

	flowIDs, changeAmount := flowIDAndAmount(user.UID())
	if changeAmount < conf.GetCfgHall().WithDrawMin {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: fee,
			Error:  msg.ErrWithDrawLack,
			ErrMsg: fmt.Sprintln("最低提奖", conf.GetCfgHall().WithDrawMin, "元"),
		})
		return
	}

	if changeAmount > config.GetCfgNormal().AmountLimit {
		changeGameWithDraw(user, changeAmount, fee, flowIDs,"", WriteFlowData)
	} else {
		if err := callWithDraw(user.BaseData.UserData.UserID, changeAmount); err != nil {
			log.Error(err.Error())
			user.WriteMsg(&msg.S2C_WithDraw{
				Amount: fee,
				Error:  msg.ErrWithDrawFail,
				ErrMsg: err.Error(),
			})
			changeGameWithDraw(user, changeAmount, fee, flowIDs,err.Error(), WriteWithdrawFinalFlowData2)
			return
		}
		user.GetUserData().IsWithdraw = true
		//ud := user.GetUserData()
		//ud.Fee -= changeAmount
		//user.WriteMsg(&msg.S2C_WithDraw{
		//	Amount: fee,
		//	Error:  msg.ErrWithDrawSuccess,
		//	ErrMsg: "成功",
		//})
		//game.GetSkeleton().Go(func() {
		//	player.SaveUserData(ud)
		//	WriteFlowData(user.UID(), changeAmount, FlowTypeWithDraw, "", "", flowIDs)
		//}, func() {
		//	sendAwardInfo(user)
		//	UpdateUserAfterTaxAward(user)
		//})
		changeGameWithDraw(user, changeAmount, fee, flowIDs,"提现成功", WriteWithdrawFinalFlowData)
	}
}

func changeGameWithDraw(user *player.User, changeAmount, fee float64, flowIDs []int,desc string , writeFlowData func(uid int, amount float64, flowType int, matchType, matchID string, flows []int,desc string)) {
	ud := user.GetUserData()
	ud.Fee -= changeAmount
	user.WriteMsg(&msg.S2C_WithDraw{
		Amount: fee,
		Error:  msg.ErrWithDrawSuccess,
		ErrMsg: "成功",
	})
	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
		writeFlowData(user.UID(), changeAmount, FlowTypeWithDraw, "", "", flowIDs,desc)
	}, func() {
		sendAwardInfo(user)
		UpdateUserAfterTaxAward(user)
	})
}

func flowIDAndAmount(id int) (flowIDs []int, changeAmount float64) {
	fd := new(FlowData)
	fd.Userid = id
	flowdatas := fd.readAllNormal()
	for _, v := range *flowdatas {
		flowIDs = append(flowIDs, v.ID)
		v.Status = FlowDataStatusAction
		v.save()
		changeAmount += v.ChangeAmount
	}
	return flowIDs, changeAmount
}

func FeeAmount(id int) (changeAmount float64) {
	fd := new(FlowData)
	fd.Userid = id
	flowdatas := fd.readAllNormal()
	for _, v := range *flowdatas {
		changeAmount += v.ChangeAmount
	}
	return changeAmount
}

func checkWithdraw(user *player.User) {
	dead := user.GetUserData().WithdrawDeadLine
	if dead < time.Now().Unix() {
		//week := time.Unix(dead, 0).Weekday()
		//dist := 0
		//if week > time.Sunday {
		//	dist = 7 - int(week)
		//}
		//if week == time.Monday || time.Unix(dead, 0).Add(time.Duration(dist+1)*24*time.Hour).Unix() <= time.Now().Unix() {
		//	user.GetUserData().SignTimes = 0
		//}

		user.GetUserData().WithdrawDeadLine = utils.OneDay0ClockTimestamp(time.Now().Add(24 * time.Hour))
		user.GetUserData().IsWithdraw = false
		player.SaveUserData(user.GetUserData())
	}
}
