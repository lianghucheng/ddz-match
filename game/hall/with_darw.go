package hall

import (
	"ddz/conf"
	"ddz/edy_api"
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
			Error: msg.ErrWithDrawFail,
		})
		return
	}

	flowIDs, changeAmount := flowIDAndAmount()

	if changeAmount < conf.GetCfgHall().WithDrawMin {
		user.WriteMsg(&msg.S2C_WithDraw{
			Error: msg.ErrWithDrawLack,
			ErrMsg:fmt.Sprintln("最低提奖",conf.GetCfgHall().WithDrawMin,"元"),
		})
		return
	}

	if err := callWithDraw(user.BaseData.UserData.UserID, changeAmount); err != nil {
		log.Error(err.Error())
		user.WriteMsg(&msg.S2C_WithDraw{
			Error: msg.ErrWithDrawFail,
			ErrMsg:"失败",
		})
		return
	}
	ud := user.GetUserData()
	ud.Fee -= changeAmount
	go func() {
		player.SaveUserData(ud)
	}()
	user.WriteMsg(&msg.S2C_WithDraw{
		Amount: changeAmount,
		Error:  msg.ErrWithDrawSuccess,
		ErrMsg: "成功",
	})
	WriteFlowData(user.BaseData.UserData.UserID, changeAmount, FlowTypeWithDraw, "", flowIDs)
	sendAwardInfo(user)
}

func flowIDAndAmount() (flowIDs []int, changeAmount float64) {
	fd:= new(FlowData)
	flowdatas := fd.readAllNormal()
	for _, v := range *flowdatas {
		flowIDs = append(flowIDs, v.ID)
		v.Status = FlowDataStatusAction
		v.save()
		changeAmount += v.ChangeAmount
	}
	return flowIDs, changeAmount
}