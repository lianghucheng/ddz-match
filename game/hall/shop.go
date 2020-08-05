package hall

import (
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"gopkg.in/mgo.v2/bson"
)

const (
	SendSingle = 1
	SendBroacast = 2
)

func SendPayAccount(user *player.User, model int) {
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
		user.WriteMsg(m)
	} else if model == 2 {
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
			PriceID: v.ID,
			Fee:     int64(v.Price),
			Name:    values.PropTypeStr[v.PropType],
			Amount:  v.GetAmount,
			GiftAmount:v.GiftAmount,
		})
	}
	return rt
}