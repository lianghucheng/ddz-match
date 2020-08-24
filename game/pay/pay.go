package pay

import (
	"ddz/config"
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

//todo: 支付下架，引导支付失败？处理方式人工处理，或程序不做拦截。
func CreateOrder(user *player.User, m *msg.C2S_CreateEdyOrder) {
	shopMerchant := db.ReadShopMerchant()
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	pm := msg.PriceItem{}

	goods := db.ReadGoodsById(m.PriceID)

	if goods.Price > shopMerchant.PayMax || goods.Price < shopMerchant.PayMin {
		user.WriteMsg(&msg.S2C_CreateEdyOrder{
			Error:  msg.ErrCreateEdyOrderOverRange,
			ErrMsg: "超出支付范围",
		})
		return
	}

	pm.GiftAmount = goods.GiftAmount
	pm.Amount = goods.GetAmount
	pm.Fee = int64(goods.Price)
	pm.Name = config.GetPropBaseConfig(goods.PropType).Name
	pm.PriceID = goods.ID
	order.Createdat = time.Now().Unix()
	order.ID, _ = db.MongoDBNextSeq("edyorder")
	order.Accountid = user.AcountID()
	order.Merchant = values.MerchantSportCentralAthketicAssociation
	order.Fee = pm.Fee
	order.Amount = pm.Amount + pm.GiftAmount

	db.Save("edyorder", order, bson.M{"_id": order.ID})

	//todo: 定义支付类型
	//payType := -1
	//if m.DefPayType == "alipay" {
	//	payType = 5
	//} else if m.DefPayType == "wxpay" {
	//	payType = 10
	//}

	merType := strconv.Itoa(shopMerchant.MerchantType)
	cfgPay := config.GetCfgPay()[merType]
	user.WriteMsg(&msg.S2C_CreateEdyOrder{
		AppID:    cfgPay.AppID,
		AppToken: cfgPay.AppToken,
		Amount:   int(pm.Fee),
		//todo: payType要修改
		PayType: cfgPay.PayType,
		//DefPayType:m.DefPayType,
		Subject:          pm.Name,
		Description:      strconv.Itoa(int(pm.Fee/100)) + pm.Name,
		OpenOrderID:      order.TradeNo,
		OpenNotifyUrl:    cfgPay.NotifyHost + cfgPay.NotifyUrl,
		CreatePaymentUrl: cfgPay.PayHost + cfgPay.CreatePaymentUrl,
	})

	//todo: 暂时的支付
	//pm := hall.GetTempPrice()[m.PriceID]
	//if user.RealName() == "" {
	//	user.WriteMsg(&msg.S2C_CreateEdyOrder{
	//		Error:  msg.ErrCreateEdyOrderNotRealAuth,
	//		ErrMsg: "未实名认证",
	//	})
	//	return
	//}
	//
	//user.WriteMsg(&msg.S2C_CreateEdyOrder{
	//	ErrMsg:   "成功",
	//	AppID:    0,
	//	AppToken: "",
	//	Amount:   int(pm.Fee),
	//	//todo: payType要修改
	//	PayType: 1,
	//	//DefPayType:m.DefPayType,
	//	Subject:          pm.Name,
	//	Description:      strconv.Itoa(int(pm.Fee/100)) + pm.Name,
	//	OpenOrderID:      "",
	//	OpenNotifyUrl:    "",
	//	CreatePaymentUrl: "",
	//})

	//若干时间后，判定为支付失败
	game.GetSkeleton().AfterFunc(5*time.Minute, func() {
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

func FakerCreateOrder(aid int, pid int) (int, string) {
	ud := player.ReadUserDataByAid(aid)
	if ud == nil {
		return 1, "不存在用户"
	}

	shopMerchant := db.ReadShopMerchant()
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	pm := msg.PriceItem{}

	goods := db.ReadGoodsById(pid)
	if goods.ID == 0 {
		return 1, "不存在商品"
	}

	if goods.Price > shopMerchant.PayMax || goods.Price < shopMerchant.PayMin {
		return msg.ErrCreateEdyOrderOverRange, "超出支付范围"
	}

	pm.GiftAmount = goods.GiftAmount
	pm.Amount = goods.GetAmount
	pm.Fee = int64(goods.Price)
	pm.Name = config.GetPropBaseConfig(goods.PropType).Name
	pm.PriceID = goods.ID
	order.Createdat = time.Now().Unix()
	order.ID, _ = db.MongoDBNextSeq("edyorder")
	order.Accountid = aid
	order.Merchant = values.MerchantSportCentralAthketicAssociation
	order.Fee = pm.Fee
	order.Amount = pm.Amount + pm.GiftAmount
	order.TradeNoReceive = "faker"

	db.Save("edyorder", order, bson.M{"_id": order.ID})

	//todo: 定义支付类型
	//payType := -1
	//if m.DefPayType == "alipay" {
	//	payType = 5
	//} else if m.DefPayType == "wxpay" {
	//	payType = 10
	//}

	//todo: 暂时的支付
	//pm := hall.GetTempPrice()[m.PriceID]
	//if user.RealName() == "" {
	//	user.WriteMsg(&msg.S2C_CreateEdyOrder{
	//		Error:  msg.ErrCreateEdyOrderNotRealAuth,
	//		ErrMsg: "未实名认证",
	//	})
	//	return
	//}
	//
	//user.WriteMsg(&msg.S2C_CreateEdyOrder{
	//	ErrMsg:   "成功",
	//	AppID:    0,
	//	AppToken: "",
	//	Amount:   int(pm.Fee),
	//	//todo: payType要修改
	//	PayType: 1,
	//	//DefPayType:m.DefPayType,
	//	Subject:          pm.Name,
	//	Description:      strconv.Itoa(int(pm.Fee/100)) + pm.Name,
	//	OpenOrderID:      "",
	//	OpenNotifyUrl:    "",
	//	CreatePaymentUrl: "",
	//})

	//若干时间后，判定为支付失败
	game.GetSkeleton().AfterFunc(5*time.Minute, func() {
		data := new(values.EdyOrder)
		db.Read("edyorder", data, bson.M{"tradeno": order.TradeNo})
		if data.PayStatus != 1 {
			data.PayStatus = 2
			db.Save("edyorder", data, bson.M{"_id": order.ID})
		}
	})

	return 0, "成功，订单号" + order.TradeNo
}
