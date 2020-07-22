package hall

import (
	"ddz/conf"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"fmt"

	"github.com/szxby/tools/log"
)

func WithDraw(user *player.User) {
	withDraw(user, edy_api.WithDrawAPI)
}

func withDraw(user *player.User, callWithDraw func(userid int, amount float64) error) {
	if callWithDraw == nil {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: user.Fee(),
			Error:  msg.ErrWithDrawFail,
		})
		return
	}

	if user.RealName() == "" && !user.IsTest() {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: user.Fee(),
			Error:  msg.ErrWithDrawNoAuth,
			ErrMsg: "未实名认证",
		})
		return
	}

	if user.BankCardNo() == "" && !user.IsTest() {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: user.Fee(),
			Error:  msg.ErrWithDrawNoBank,
			ErrMsg: "未绑定银行卡",
		})
		return
	}

	flowIDs, changeAmount := flowIDAndAmount(user.UID())
	if changeAmount < conf.GetCfgHall().WithDrawMin {
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: user.Fee(),
			Error:  msg.ErrWithDrawLack,
			ErrMsg: fmt.Sprintln("最低提奖", conf.GetCfgHall().WithDrawMin, "元"),
		})
		return
	}

	if err := callWithDraw(user.BaseData.UserData.UserID, changeAmount); err != nil {
		log.Error(err.Error())
		user.WriteMsg(&msg.S2C_WithDraw{
			Amount: user.Fee(),
			Error:  msg.ErrWithDrawFail,
			ErrMsg: "三方接口失败",
		})
		return
	}
	ud := user.GetUserData()
	ud.Fee -= changeAmount
	user.WriteMsg(&msg.S2C_WithDraw{
		Amount: user.Fee(),
		Error:  msg.ErrWithDrawSuccess,
		ErrMsg: "成功",
	})
	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
		WriteFlowData(user.UID(), changeAmount, FlowTypeWithDraw, "", "", flowIDs)
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