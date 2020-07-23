package pay

import (
	"ddz/config"
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

const createPaymentUrl = "https://open.test.boai1986.cn"

func CreateOrder(user *player.User, priceID int) {
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	pm := config.PriceItem{}
	for _, v := range *config.GetPriceMenu() {
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
		AppID:            edy_api.AppID,
		AppToken:         edy_api.AppToken,
		Amount:           int(order.Fee),
		PayType:          5,
		Subject:          pm.Name,
		Description:      strconv.Itoa(int(order.Fee/100)) + pm.Name,
		OpenOrderID:      order.TradeNo,
		OpenNotifyUrl:    "http://123.207.12.67:9084" + edy_api.EdyBackCall,
		CreatePaymentUrl: createPaymentUrl + "/api/payment/create",
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
