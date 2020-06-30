package edy_api

import (
	"ddz/edy_api/internal"
)

func RealAuthApi(accountid int, idCardNo, realName, phoneNum string) error {
	idBind := internal.NewIDBindReq(accountid, idCardNo, realName, phoneNum)
	return idBind.IdCardBind()
}

func BandBankCardAPI(accountid int, bankNo, BankName, BankAccount string) error {
	bindBankCard := internal.NewBindBankCardReq(accountid, bankNo, BankName, BankAccount)
	return bindBankCard.BindBankCard()
}
