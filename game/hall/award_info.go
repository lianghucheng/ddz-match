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
		WithDrawList: *withDrawList(flowData.readAllByUserID()),
	})
}

func withDrawList(flowDatas *[]FlowData) *[]msg.WithDrawData {
	rt := new([]msg.WithDrawData)
	for _, v := range *flowDatas {
		status := ""
		if v.FlowType == FlowTypeAward {
			status = FlowDataStatusMsg[FlowDataStatusNormal]
		}else if v.FlowType == FlowTypeGift {
			status = FlowDataStatusMsg[FlowDataStatusGift]
		} else {
			status = FlowDataStatusMsg[v.Status]
		}
		*rt = append(*rt, msg.WithDrawData{
			FlowType:  v.FlowType,
			MatchID:   v.MatchID,
			Amount:    v.ChangeAmount,
			Status:    status,
			CreatedAt: v.CreatedAt,
		})
	}

	return rt
}
