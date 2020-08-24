package hall

import (
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/lianhang_api"
	"ddz/msg"
	"errors"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

type BankCard struct {
	Userid      int
	BankName    string
	BankCardNo  string
	Province    string
	City        string
	OpeningBank string
	OpeningBankNo string
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
		SendAddBankCard(user, msg.ErrAddBankCardFail, "绑定失败")
		return
	}
	//if user.BankCardNo() != "" {
	//	SendAddBankCard(user, msg.ErrAddBankCardAlready, "重复绑定")
	//	return
	//}
	var err error
	game.GetSkeleton().Go(func() {
		ud := player.ReadUserDataByID(ctx.Userid)
		lianHangReq := new(lianhang_api.LianHangReq)
		lianHangReq.Bank = ctx.BankName
		lianHangReq.Bankcard = ctx.BankCardNo
		lianHangReq.City = ctx.City
		lianHangReq.Key = ctx.OpeningBank
		lianHangReq.Province = ctx.Province
		bankCode := ""
		bankCode, err = lianHangReq.LianHangApi()
		if err != nil {
			log.Error(err.Error())
			err = errors.New("查询不到联行号，请联系客服解决～")
			return
		}
		if bankCode == "" {
			err = errors.New("查询不到联行号，请联系客服解决～")
			return
		}
		ctx.OpeningBankNo = bankCode
		aid := ud.AccountID
		err = api(aid, bankCode, ctx.BankName, ctx.BankCardNo)
	}, func() {
		if err != nil {
			SendAddBankCard(user, msg.ErrAddBankCardBusiness, err.Error())
			return
		}
		user.BaseData.UserData.BankCardNo = ctx.BankCardNo
		player.SaveUserData(user.GetUserData())
		err = ctx.save()
		if err != nil {
			SendAddBankCard(user, msg.ErrAddBankCardFail, "绑定失败")
			return
		}
		SendAddBankCard(user, msg.ErrAddBankCardSuccess, "绑定成功")
	})
}
