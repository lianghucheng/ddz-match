package hall

import (
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type BankCard struct {
	Userid      int
	BankName    string
	BankCardNo  string
	Province    string
	OpeningBank string
}

func (ctx *BankCard) save() error {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	if _, err := se.DB(db.DB).C("bankcard").Upsert(bson.M{"userid": ctx.Userid}, ctx); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (ctx *BankCard) read() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	if err := se.DB(db.DB).C("bankcard").Find(bson.M{"userid": ctx.Userid}).One(ctx); err != nil {
		log.Error(err.Error())
	}
}

func AddBankCard(user *player.User, m *msg.C2S_AddBankCard) {
	addBankCard(user, m, nil)
}

func addBankCard(user *player.User, m *msg.C2S_AddBankCard, cb func()) {
	bankCard := new(BankCard)
	bankCard.Userid = user.BaseData.UserData.UserID
	bankCard.BankName = m.BankName
	bankCard.BankCardNo = m.BankCardNo
	bankCard.Province = m.Province
	bankCard.OpeningBank = m.OpeningBank

	if err := bankCard.save(); err != nil {
		user.WriteMsg(&msg.S2C_AddBankCard{
			Error: msg.ErrAddBankCardFail,
		})
		return
	}
	user.WriteMsg(&msg.S2C_AddBankCard{
		Error: msg.ErrAddBankCardSuccess,
	})
	if cb != nil {
		cb()
	}
}
