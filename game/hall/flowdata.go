package hall

import (
	"ddz/game"
	"ddz/game/db"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	FlowTypeAward    = 1
	FlowTypeWithDraw = 2
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
	ChangeAmount    		float64
	FlowType  				int
	MatchType 				string
	Status    				int
	CreatedAt 				int64
	FlowID 					[]int
}

func (ctx *FlowData) save() {
	game.GetSkeleton().Go(func() {
		se := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(se)
		_, err := se.DB(db.DB).C("flowdata").Upsert(bson.M{"_id": ctx.ID}, ctx)
		if err != nil {
			log.Error(err.Error())
		}
	}, nil)
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
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"status": FlowDataStatusNormal}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func WriteFlowData(userid int, amount float64, flowType int, matchType string, flows []int) {
	flowData := new(FlowData)
	flowData.Userid = userid
	flowData.ChangeAmount = amount
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.CreatedAt = time.Now().Unix()
	flowData.FlowID = flows
	flowData.ID, _ = db.MongoDBNextSeq("flowdata")
	flowData.save()
}
