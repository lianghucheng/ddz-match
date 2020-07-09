package edy_api

import (
	"ddz/edy_api/internal"
)

func RealAuthApi(accountid int, idCardNo, realName, phoneNum string) error {
	idBind := internal.NewIDBindReq(accountid, idCardNo, realName, phoneNum)
	return idBind.IdCardBind()
}

func RealAuthApi2(accountid int, idCardNo, realName, phoneNum string) error {
	return nil
}

func BandBankCardAPI(accountid int, bankNo, BankName, BankAccount string) error {
	bindBankCard := internal.NewBindBankCardReq(accountid, bankNo, BankName, BankAccount)
	return bindBankCard.BindBankCard()
}

func BandBankCardAPI2(accountid int, bankNo, BankName, BankAccount string) error {
	return nil
}

func WithDrawAPI(userid int, amount float64) error {
	return nil
}
