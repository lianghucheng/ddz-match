package hall

import (
	"ddz/config"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	SendSingle   = 1
	SendBroacast = 2
)

func SendPayAccount(user *player.User, model int) {
	log.Debug("send pay account")
	mer := db.ReadShopMerchant()
	payAccounts := []string{}
	for _, v := range mer.UpPayBranchs {
		pa := db.ReadPayAccounts(mer.ID, v)
		for _, v := range *pa {
			payAccounts = append(payAccounts, v.Account)
		}
	}
	m := &msg.S2C_PayAccount{
		Accounts: payAccounts,
	}
	if model == 1 {
		log.Debug("single %v", m)
		user.WriteMsg(m)
	} else if model == 2 {
		log.Debug("broadcast %v", m)
		player.Broadcast(m)
	}
}

func GetPriceMenu(goodsTypeID int) *[]msg.PriceItem {
	rt := new([]msg.PriceItem)
	//*rt = append(*rt, PriceItem{
	//	PriceID: 1,
	//	Fee:     20000,
	//	Name:    "点券",
	//	Amount:  200,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 2,
	//	Fee:     10000,
	//	Name:    "点券",
	//	Amount:  100,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 3,
	//	Fee:     5000,
	//	Name:    "点券",
	//	Amount:  50,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 4,
	//	Fee:     2000,
	//	Name:    "点券",
	//	Amount:  20,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 5,
	//	Fee:     1000,
	//	Name:    "点券",
	//	Amount:  10,
	//})
	//*rt = append(*rt, PriceItem{
	//	PriceID: 6,
	//	Fee:     500,
	//	Name:    "点券",
	//	Amount:  5,
	//})

	goodses := db.ReadGoodses(bson.M{"goodstypeid": goodsTypeID})
	for _, v := range *goodses {
		*rt = append(*rt, msg.PriceItem{
			PriceID:    v.ID,
			Fee:        int64(v.Price),
			Name:       config.GetPropBaseConfig(v.PropType).Name,
			Amount:     v.GetAmount,
			GiftAmount: v.GiftAmount,
			ImgUrl:     config.GetPropBaseConfig(v.PropType).ImgUrl,
			TakenType:  v.TakenType,
		})
	}
	return rt
}

func SendPriceMenu(user *player.User, model int) {
	log.Debug("send price menu")
	//merchant := db.ReadShopMerchant()
	//if merchant.ID <= 0 {
	//	log.Error("Has no up merchant int shop. ")
	//	return
	//}
	//goodsTypes := db.ReadGoodsTypes(merchant.ID)
	//if len(*goodsTypes) == 0 {
	//	log.Error("The goodsType is nil. ")
	//	return
	//}
	//
	//if len(*goodsTypes) == 0 {
	//	log.Error("The goodsType is nil. ")
	//	return
	//}
	//msgGoodsTypes := new([]msg.GoodsType)
	//for _, v := range *goodsTypes {
	//	*msgGoodsTypes = append(*msgGoodsTypes, msg.GoodsType{
	//		ID:         v.ID,
	//		TypeName:   v.TypeName,
	//		ImgUrl:     v.ImgUrl,
	//		PriceItems: *GetPriceMenu(v.ID),
	//	})
	//}
	//m := &msg.S2C_PriceMenu{
	//	PriceItems: *msgGoodsTypes,
	//}

	cf := config.GetPropBaseConfig
	log.Debug("获取到的缓存道具基本信息：%v", cf(values.PropTypeCoupon))

	prices := []msg.PriceItem{
		{
			PriceID:    1,
			Fee:        500,
			Name:       cf(values.PropTypeCoupon).Name,
			Amount:     5,
			ImgUrl:     cf(values.PropTypeCoupon).ImgUrl,
			TakenType:  1,
			GiftAmount: 0,
			//Name:   "点券",
			//ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/juan.png",
		},
		{
			PriceID:    2,
			Fee:        1000,
			Name:       cf(values.PropTypeCoupon).Name,
			Amount:     10,
			ImgUrl:     cf(values.PropTypeCoupon).ImgUrl,
			TakenType:  1,
			GiftAmount: 0,
			//Name:   "点券",
			//ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/juan.png",
		},
		{
			PriceID:    3,
			Fee:        2000,
			Name:       cf(values.PropTypeCoupon).Name,
			Amount:     20,
			ImgUrl:     cf(values.PropTypeCoupon).ImgUrl,
			TakenType:  1,
			GiftAmount: 0,
			//Name:   "点券",
			//ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/juan.png",
		},
		{
			PriceID:    4,
			Fee:        5000,
			Name:       cf(values.PropTypeCoupon).Name,
			Amount:     50,
			ImgUrl:     cf(values.PropTypeCoupon).ImgUrl,
			TakenType:  1,
			GiftAmount: 0,
			//Name:   "点券",
			//ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/juan.png",
		},
		{
			PriceID:    5,
			Fee:        10000,
			Name:       cf(values.PropTypeCoupon).Name,
			Amount:     100,
			ImgUrl:     cf(values.PropTypeCoupon).ImgUrl,
			TakenType:  1,
			GiftAmount: 0,
			//Name:   "点券",
			//ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/juan.png",
		},
	}
	data := []msg.GoodsType{
		{
			ID:         0,
			TypeName:   "点券",
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_dianquan.png",
			PriceItems: prices,
		},
	}
	m := &msg.S2C_PriceMenu{
		PriceItems: data,
	}
	if model == 1 {
		log.Debug("price menu single %v   %v", *m, (m.PriceItems))
		user.WriteMsg(m)
	} else if model == 2 {
		log.Debug("price menu broadcast %v   %v", *m, (m.PriceItems))
		player.Broadcast(m)
	}
}

func GetTempPrice() map[int]msg.PriceItem {
	data := map[int]msg.PriceItem{
		1: {
			PriceID:    1,
			Fee:        500,
			Name:       "点券",
			Amount:     5,
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_juan.png",
			TakenType:  1,
			GiftAmount: 0,
		},
		2: {
			PriceID:    2,
			Fee:        1000,
			Name:       "点券",
			Amount:     10,
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_juan.png",
			TakenType:  1,
			GiftAmount: 0,
		},
		3: {
			PriceID:    3,
			Fee:        2000,
			Name:       "点券",
			Amount:     20,
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_juan.png",
			TakenType:  1,
			GiftAmount: 0,
		},
		4: {
			PriceID:    4,
			Fee:        5000,
			Name:       "点券",
			Amount:     50,
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_juan.png",
			TakenType:  1,
			GiftAmount: 0,
		},
		5: {
			PriceID:    5,
			Fee:        10000,
			Name:       "点券",
			Amount:     100,
			ImgUrl:     "http://111.230.39.198:10615/download/matchIcon/bg_juan.png",
			TakenType:  1,
			GiftAmount: 0,
		},
	}
	return data
}
