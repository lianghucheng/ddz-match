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
			order.Amount = v.Amount
			break
		}
	}
	order.Createdat = time.Now().Unix()
	order.ID, _ = db.MongoDBNextSeq("edyorder")
	order.Accountid = user.AcountID()
	order.Merchant = values.MerchantSportCentralAthketicAssociation
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

	//若干时间后，判定为支付失败
	game.GetSkeleton().AfterFunc(5*time.Minute, func(){
		data := new(values.EdyOrder)
		db.Read("edyorder", data, bson.M{"tradeno": order.TradeNo})
		if data.PayStatus != 1 {
			data.PayStatus = 2
			db.Save("edyorder", data, bson.M{"_id": order.ID})
		}
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
