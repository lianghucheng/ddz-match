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

func CreateOrder(user *player.User, m *msg.C2S_CreateEdyOrder) {
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	pm := config.PriceItem{}
	for _, v := range *config.GetPriceMenu() {
		if v.PriceID == m.PriceID {
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
	//payType := -1
	//if m.DefPayType == "alipay" {
	//	payType = 5
	//} else if m.DefPayType == "wxpay" {
	//	payType = 10
	//}

	user.WriteMsg(&msg.S2C_CreateEdyOrder{
		AppID:            edy_api.AppID,
		AppToken:         edy_api.AppToken,
		Amount:           int(order.Fee),
		PayType:          9,
		//DefPayType:m.DefPayType,
		Subject:          pm.Name,
		Description:      strconv.Itoa(int(order.Fee/100)) + pm.Name,
		OpenOrderID:      order.TradeNo,
		OpenNotifyUrl:    config.GetCfgPay().Host + edy_api.EdyBackCall,
		CreatePaymentUrl: config.GetCfgPay().CreatePaymentUrl + "/api/payment/create",
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
