package hall

import (
	"ddz/game"
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
	City        string
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
	bankCard := new(BankCard)
	bankCard.Userid = user.BaseData.UserData.UserID
	bankCard.BankName = m.BankName
	bankCard.BankCardNo = m.BankCardNo
	bankCard.Province = m.Province
	bankCard.City = m.City
	bankCard.OpeningBank = m.OpeningBank
	bankCard.addBankCard(user, m, BandBankCardAPI)
}

func (ctx *BankCard)addBankCard(user *player.User, m *msg.C2S_AddBankCard, cb func(b *BankCard) error) {
	if cb == nil {
		SendAddBankCard(user, msg.ErrAddBankCardFail)
		return
	}
	if user.BankCardNo() != "" {
		SendAddBankCard(user, msg.ErrAddBankCardAlready)
		return
	}
	var err error
	game.GetSkeleton().Go(func() {
		err = cb(ctx)
	}, func() {
		if err != nil {
			SendAddBankCard(user, msg.ErrAddBankCardFail)
			return
		}
		user.BaseData.UserData.BankCardNo = ctx.BankCardNo
		player.SaveUserData(user.GetUserData())
		err = ctx.save()
		if err != nil {
			SendAddBankCard(user, msg.ErrAddBankCardFail)
			return
		}
		SendAddBankCard(user, msg.ErrAddBankCardSuccess)
	})
}

func BandBankCardAPI(b *BankCard) error {
	return nil
}
