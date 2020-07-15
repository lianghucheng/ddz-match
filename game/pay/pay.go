package pay

import (
	"ddz/conf2"
	"ddz/edy_api"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func CreateOrder(user *player.User, priceID int) {
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	for _, v := range *conf2.GetPriceMenu() {
		if v.PriceID == priceID {
			order.Fee = v.Fee
			break
		}
	}
	order.Createdat = time.Now().Unix()
	order.ID, _ = db.MongoDBNextSeq("edyorder")
	order.Accountid = user.AcountID()
	db.Save("edyorder", order, bson.M{"_id": order.ID})
	user.WriteMsg(&msg.S2C_CreateEdyOrder{
		TradeNo:order.TradeNo,
		NotifyUrl:edy_api.EdyBackCall,
	})
}