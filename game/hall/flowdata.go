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

type FlowData struct {
	Userid    int
	Amount    float64
	FlowType  int
	MatchType string
	CreatedAt int64
}

func (ctx *FlowData) save() {
	game.GetSkeleton().Go(func() {
		se := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(se)
		_, err := se.DB(db.DB).C("flowdata").Upsert(bson.M{"userid": ctx.Userid}, ctx)
		if err != nil {
			log.Error(err.Error())
		}
	}, nil)
}

func (ctx *FlowData) readAllByID() *[]FlowData {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	rt := new([]FlowData)
	err := se.DB(db.DB).C("flowdata").Find(bson.M{"userid": ctx.Userid}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}

	return rt
}

func WriteFlowData(userid int, amount float64, flowType int, matchType string) {
	flowData := new(FlowData)
	flowData.Userid = userid
	flowData.Amount = amount
	flowData.FlowType = flowType
	flowData.MatchType = matchType
	flowData.CreatedAt = time.Now().Unix()

	flowData.save()
}
