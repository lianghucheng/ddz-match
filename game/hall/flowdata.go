package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	FlowTypeAward    = 1
	FlowTypeWithDraw = 2
	FlowTypeGift     = 3
	FlowTypeSign     = 4
)

const (
	FlowDataStatusNormal = 0
	FlowDataStatusAction = 1
	FlowDataStatusOver   = 2
	FlowDataStatusBack   = 3
	FlowDataStatusGift   = 4
	FlowDataStatusSign   = 5
)

var FlowDataStatusMsg = map[int]string{
	FlowDataStatusNormal: "比赛获得",
	FlowDataStatusAction: "提奖中",
	FlowDataStatusOver:   "已提奖",
	FlowDataStatusBack:   "已退奖",
	FlowDataStatusGift:   "平台赠送",
}

type FlowData struct {
	ID           int `bson:"_id"`
	Userid       int
	Accountid    int
	ChangeAmount float64
	FlowType     int
	MatchType    string
	MatchID      string
	Status       int
	CreatedAt    int64
	FlowIDs      []int
	Realname     string
	TakenFee     float64
	AtferTaxFee  float64
	Desc         string
	PassStatus   int //1是已通过，0是未通过
	ActMoney     string
}

func (ctx *FlowData) save() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	_, err := se.DB(db.DB).C("flowdata").Upsert(bson.M{"_id": ctx.ID}, ctx)
	if err != nil {
		log.Error(err.Error())
	}
}

func (ctx *FlowData) readAllByID() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid": ctx.ID}).One(ctx)
	if err != nil {
		log.Error(err.Error())
	}
}

func (ctx *FlowData) readAllByUserID() *[]FlowData {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	rt := new([]FlowData)
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid": ctx.Userid}).Sort("-createdat").Limit(40).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func (ctx *FlowData) readAllNormal() *[]FlowData {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	rt := new([]FlowData)
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid": ctx.Userid, "status": FlowDataStatusNormal}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func WriteFlowData(uid int, amount float64, flowType int, matchType, matchID string, flows []int, data map[string]interface{}) {
	log.Debug("奖金流水数据变动：uid: %v, amount: %v, flowType: %v, matchType: %v, matchID: %v, flows: %v. ", uid, amount, flowType, matchType, matchID, flows)
	ud := player.ReadUserDataByID(uid)
	flowData := new(FlowData)
	flowData.Userid = ud.UserID

	flowData.ChangeAmount = utils.Decimal(amount)
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.MatchID = matchID
	flowData.CreatedAt = time.Now().Unix()
	flowData.FlowIDs = flows
	flowData.Realname = ud.RealName
	flowData.TakenFee = ud.TakenFee
	flowData.AtferTaxFee = ud.Fee
	flowData.Accountid = ud.AccountID
	flowData.ID, _ = db.MongoDBNextSeq("flowdata")
	if flowType == FlowTypeWithDraw {
		flowData.Status = FlowDataStatusAction
	}
	flowData.save()
	game.GetSkeleton().ChanRPCServer.Go("UpdateAwardInfo", &msg.RPC_UpdateAwardInfo{
		Uid: uid,
	})
}

func WriteFlowDataWithTime(uid int, amount float64, flowType int, matchType, matchID string, flows []int, timestamp int64, status int) int {
	log.Debug("奖金流水数据变动：uid: %v, amount: %v, flowType: %v, matchType: %v, matchID: %v, flows: %v. ", uid, amount, flowType, matchType, matchID, flows)
	ud := player.ReadUserDataByID(uid)
	flowData := new(FlowData)
	flowData.Userid = ud.UserID

	flowData.ChangeAmount = utils.Decimal(amount)
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.MatchID = matchID
	flowData.CreatedAt = timestamp
	flowData.FlowIDs = flows
	flowData.Realname = ud.RealName
	flowData.TakenFee = ud.TakenFee
	flowData.AtferTaxFee = ud.Fee
	flowData.Accountid = ud.AccountID
	flowData.ID, _ = db.MongoDBNextSeq("flowdata")
	if flowType == FlowTypeWithDraw {
		flowData.Status = FlowDataStatusAction
	}
	flowData.Status=status
	flowData.save()
	game.GetSkeleton().ChanRPCServer.Go("UpdateAwardInfo", &msg.RPC_UpdateAwardInfo{
		Uid: uid,
	})

	return flowData.ID
}

func WriteWithdrawFinalFlowData(uid int, amount float64, flowType int, matchType, matchID string, flows []int, data map[string]interface{}) {
	log.Debug("奖金流水数据变动：uid: %v, amount: %v, flowType: %v, matchType: %v, matchID: %v, flows: %v. ", uid, amount, flowType, matchType, matchID, flows)
	ud := player.ReadUserDataByID(uid)
	flowData := new(FlowData)
	flowData.Userid = ud.UserID

	flowData.ChangeAmount = utils.Decimal(amount)
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.MatchID = matchID
	flowData.CreatedAt = time.Now().Unix()
	flowData.FlowIDs = flows
	flowData.Realname = ud.RealName
	flowData.TakenFee = ud.TakenFee
	flowData.AtferTaxFee = ud.Fee
	flowData.Accountid = ud.AccountID
	flowData.PassStatus = 1
	flowData.Desc = data["resp_msg"].(string)
	flowData.ActMoney = data["act_money"].(string)
	flowData.ID, _ = db.MongoDBNextSeq("flowdata")
	flowData.Status = FlowDataStatusOver
	flowData.save()
	paymentByFlowIDs(flows)
	game.GetSkeleton().ChanRPCServer.Go("UpdateAwardInfo", &msg.RPC_UpdateAwardInfo{
		Uid: uid,
	})
}

func WriteWithdrawFinalFlowData2(uid int, amount float64, flowType int, matchType, matchID string, flows []int, data map[string]interface{}) {
	log.Debug("奖金流水数据变动：uid: %v, amount: %v, flowType: %v, matchType: %v, matchID: %v, flows: %v. ", uid, amount, flowType, matchType, matchID, flows)
	ud := player.ReadUserDataByID(uid)
	flowData := new(FlowData)
	flowData.Userid = ud.UserID

	flowData.ChangeAmount = utils.Decimal(amount)
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.MatchID = matchID
	flowData.CreatedAt = time.Now().Unix()
	flowData.FlowIDs = flows
	flowData.Realname = ud.RealName
	flowData.TakenFee = ud.TakenFee
	flowData.AtferTaxFee = ud.Fee
	flowData.Accountid = ud.AccountID
	flowData.PassStatus = 1
	flowData.Desc = data["resp_msg"].(string)
	flowData.ID, _ = db.MongoDBNextSeq("flowdata")
	flowData.Status = FlowDataStatusBack
	flowData.save()
	refundByFlowIDs(flows)
	game.GetSkeleton().ChanRPCServer.Go("UpdateAwardInfo", &msg.RPC_UpdateAwardInfo{
		Uid: uid,
	})
}

func paymentByFlowIDs(flowIDs []int) {
	for _, v := range flowIDs {
		fd := db.ReadFlowDataByID(v)
		fd.Status = FlowDataStatusOver
		data := new(FlowData)
		if err := utils.Transfer(fd, data); err != nil {
			log.Error(err.Error())
		}
		data.save()
	}
}

func refundByFlowIDs(flowIDs []int) {
	for _, v := range flowIDs {
		fd := db.ReadFlowDataByID(v)
		fd.Status = FlowDataStatusNormal
		data := new(FlowData)
		if err := utils.Transfer(fd, data); err != nil {
			log.Error(err.Error())
		}
		data.save()
	}
}
