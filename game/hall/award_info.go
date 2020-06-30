package hall

import (
	"ddz/game/player"
	"ddz/msg"
)

func SendAwardInfo(user *player.User) {
	sendAwardInfo(user)
}

func sendAwardInfo(user *player.User) {
	flowData := new(FlowData)
	flowData.Userid = user.BaseData.UserData.UserID
	user.WriteMsg(&msg.S2C_AwardInfo{
		Amount:       user.Fee(), //todo:尚未开发
		WithDrawList: *withDrawList(flowData.readAllByID()),
	})
}

func withDrawList(flowDatas *[]FlowData) *[]msg.WithDrawData {
	rt := new([]msg.WithDrawData)
	for _, v := range *flowDatas {
		*rt = append(*rt, msg.WithDrawData{
			FlowType:  v.FlowType,
			MatchType: v.MatchType,
			Amount:    v.Amount,
			Status:    v.Status,
			CreatedAt: v.CreatedAt,
		})
	}

	return rt
}
