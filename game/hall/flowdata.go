package hall

import (
	"ddz/game/db"
	"ddz/game/player"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	FlowTypeAward    = 1
	FlowTypeWithDraw = 2
	FlowTypeGift     = 3
)

const (
	FlowDataStatusNormal = 0
	FlowDataStatusAction = 1
	FlowDataStatusOver   = 2
	FlowDataStatusBack   = 3
)

type FlowData struct {
	ID    					int `bson:"_id"`
	Userid    				int
	Accountid      			int
	ChangeAmount    		float64
	FlowType  				int
	MatchType 				string
	MatchID 				string
	Status    				int
	CreatedAt 				int64
	FlowIDs 				[]int
	Realname 				string
	TakenFee 			float64
	AtferTaxFee 				float64
	Desc 					string
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
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid": ctx.Userid}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func (ctx *FlowData) readAllNormal() *[]FlowData {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	rt := new([]FlowData)
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid":ctx.Userid ,"status": FlowDataStatusNormal}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func WriteFlowData(uid int, amount float64, flowType int, matchType,matchID string, flows []int) {
	ud := player.ReadUserDataByID(uid)
	flowData := new(FlowData)
	flowData.Userid = ud.UserID
	flowData.ChangeAmount = amount
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
}