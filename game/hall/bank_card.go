package hall

import (
	"ddz/edy_api"
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

func (ctx *BankCard) Read() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	if err := se.DB(db.DB).C("bankcard").Find(bson.M{"userid": ctx.Userid}).One(ctx); err != nil {
		log.Error(err.Error())
	}
}

func AddBankCard(user *player.User, m *msg.C2S_BindBankCard) {
	bankCard := new(BankCard)
	bankCard.Userid = user.BaseData.UserData.UserID
	bankCard.BankName = m.BankName
	bankCard.BankCardNo = m.BankCardNo
	bankCard.Province = m.Province
	bankCard.City = m.City
	bankCard.OpeningBank = m.OpeningBank
	bankCard.addBankCard(user, edy_api.BandBankCardAPI)
}

func (ctx *BankCard) addBankCard(user *player.User, api func(accountid int, bankNo, BankName, BankAccount string) error) {
	if api == nil {
		SendAddBankCard(user, msg.ErrAddBankCardFail)
		return
	}
	if user.BankCardNo() != "" {
		SendAddBankCard(user, msg.ErrAddBankCardAlready)
		return
	}
	var err error
	game.GetSkeleton().Go(func() {
		err = api(ctx.Userid, ctx.OpeningBank, ctx.BankName, ctx.BankCardNo)
	}, func() {
		if err != nil {
			SendAddBankCard(user, msg.ErrAddBankCardBusiness)
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
