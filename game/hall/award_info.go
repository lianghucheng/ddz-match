package hall

import (
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
)

func SendAwardInfo(user *player.User) {
	sendAwardInfo(user)
}

func sendAwardInfo(user *player.User) {
	flowData := new(FlowData)
	flowData.Userid = user.BaseData.UserData.UserID
	changeAmount := FeeAmount(user.UID())
	user.GetUserData().Fee = changeAmount
	game.GetSkeleton().Go(func() {
		player.SaveUserData(user.GetUserData())
	}, nil)
	user.WriteMsg(&msg.S2C_AwardInfo{
		Amount:       utils.Decimal(changeAmount),
		WithDrawList: *withDrawList(flowData.readAllByUserID()),
	})
}

func withDrawList(flowDatas *[]FlowData) *[]msg.WithDrawData {
	rt := new([]msg.WithDrawData)
	for _, v := range *flowDatas {
		status := FlowDataStatusMsg[v.Status]
		if v.FlowType == FlowTypeAward {
			status = FlowDataStatusMsg[FlowDataStatusNormal]
		} else if v.FlowType == FlowTypeGift {
			status = FlowDataStatusMsg[FlowDataStatusGift]
		} else {
			status = FlowDataStatusMsg[v.Status]
		}
		matchID := v.MatchID
		if v.FlowType == FlowTypeAward {

		} else if v.Status == FlowDataStatusAction {
			matchID = "平台审核中,请稍后"
		} else if v.Status == FlowDataStatusOver {
			matchID = "提奖成功"
		} else if v.Status == FlowDataStatusBack {
			matchID = v.Desc
		}
		*rt = append(*rt, msg.WithDrawData{
			FlowType:  v.FlowType,
			MatchID:   matchID,
			Amount:    v.ChangeAmount,
			Status:    status,
			CreatedAt: v.CreatedAt,
		})
	}

	return rt
}
