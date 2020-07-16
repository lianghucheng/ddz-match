package pay

import (
	"ddz/conf2"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

const (
	appID = 100001
	appToken = "fddda32b4cb543babbf78a4ba955c05d"
	appSecret = "fddda32b4cb543babbf78a4ba955c05d"
)

func CreateOrder(user *player.User, priceID int) {
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	pm := conf2.PriceItem{}
	for _, v := range *conf2.GetPriceMenu() {
		if v.PriceID == priceID {
			pm = v
			order.Fee = v.Fee
			break
		}
	}
	order.Createdat = time.Now().Unix()
	order.ID, _ = db.MongoDBNextSeq("edyorder")
	order.Accountid = user.AcountID()
	db.Save("edyorder", order, bson.M{"_id": order.ID})
	user.WriteMsg(&msg.S2C_CreateEdyOrder{
		AppID:appID,
		AppToken:appToken,
		Amount:int(order.Fee),
		PayType:5,
		Subject:pm.Name,
		Description:strconv.Itoa(int(order.Fee/ 100))+pm.Name,
		OpenOrderID:order.TradeNo,
		OpenNotifyUrl:"http://123.207.12.67:9085"+edy_api.EdyBackCall,
	})
}

func CreateOrderSuccess(user *player.User, m *msg.C2S_CreateOrderSuccess) {
	order := new(values.EdyOrder)
	game.GetSkeleton().Go(func() {
		db.Read("edyorder", order, bson.M{"tradeno": m.OpenOrderID})
	}, func() {
		order.TradeNoReceive = m.OrderID
		db.Save("edyorder", order, bson.M{"_id": order.ID})
	})
}