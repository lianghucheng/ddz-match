package hall

import (
	"ddz/game/player"
	"ddz/msg"

	"github.com/szxby/tools/log"
)

func WithDraw(user *player.User, amount float64) {
	withDraw(user, amount, nil)
}

func withDraw(user *player.User, amount float64, callWithDraw func(userid int, amount float64) error) {
	if callWithDraw == nil {
		user.WriteMsg(&msg.S2C_WithDraw{
			Error: msg.ErrWithDrawFail,
		})
		return
	}
	if err := callWithDraw(user.BaseData.UserData.UserID, amount); err != nil {
		log.Error(err.Error())
		user.WriteMsg(&msg.S2C_WithDraw{
			Error: msg.ErrWithDrawFail,
		})
		return
	}
	user.WriteMsg(&msg.S2C_WithDraw{
		Amount: amount,
		Error:  msg.ErrWithDrawSuccess,
	})
	WriteFlowData(user.BaseData.UserData.UserID, amount, FlowTypeWithDraw, "")
	sendAwardInfo(user)
}
