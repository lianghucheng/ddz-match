package edy_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/name5566/leaf/log"
)

var (
	bind_id_uri   = "/player/identity_number/bind"
	bind_bank_uri = "/player/bank_acc/bind"
)

func RealAuthApi(accountid int, idCardNo, realName, phoneNum string) error {
	idBind := NewIDBindReq(accountid, idCardNo, realName, phoneNum)
	return idBind.idCardBind()
}

func (ctx *IDBindReq) idCardBind() error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return err
	}

	c := NewClient(bind_id_uri, string(b), reqPost)
	c.GenerateSign(reqPost)
	rt, err := c.DoPost()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	fmt.Println("【绑定身份证】", string(rt))
	res := new(IDBindResp)
	if err := json.Unmarshal(rt, res); err != nil {
		return err
	}

	if res.RespCode != "000000" {
		log.Error("【返回的错误码】%v", res.RespCode)
		return errors.New(ErrMsg[res.RespCode])
	}

	return nil
}

func BandBankCardAPI(accountid int, bankNo, BankName, BankAccount string) error {
	bindBankCard := NewBindBankCardReq(accountid, bankNo, BankName, BankAccount)
	return bindBankCard.BindBankCard()
}

func (ctx *BindBankCardReq) BindBankCard() error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return err
	}

	c := NewClient(bind_bank_uri, string(b), reqPost)
	c.GenerateSign(reqPost)
	rt, err := c.DoPost()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	fmt.Println("【绑定银行卡】", string(rt))
	res := new(BindBankCardResp)
	if err := json.Unmarshal(rt, res); err != nil {
		return err
	}

	if res.RespCode != "000000" {
		log.Error("【返回的错误码】%v", res.RespCode)
		return errors.New(ErrMsg[res.RespCode])
	}

	return nil
}
