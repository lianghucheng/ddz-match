package internal

import (
	"ddz/edy_api/internal/base"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/szxby/tools/log"
	"strconv"
)

var (
	bind_id_uri   = "/player/identity_number/bind"
)

type BindBankCardReq struct {
	CpID        string `json:"cp_id"`        //厂商ID，必填
	PlayerID    string `json:"player_id"`    //厂商端用户ID，必填
	BankNo      string `json:"bank_no"`      //开户行联行号，必填
	BankName    string `json:"bank_name"`    //开户行名称 ，必填
	BankAccount string `json:"bank_account"` //银行账号 ，必填
}

type BindBankCardResp struct {
	RespCode string `json:"resp_code"` //编码，Y
	RespMsg  string `json:"resp_msg"`  //信息，Y
}

func NewBindBankCardReq(accountid int, bankNo, BankName, BankAccount string) *BindBankCardReq {
	bindBankCard := new(BindBankCardReq)
	bindBankCard.CpID = base.CpID
	bindBankCard.PlayerID = strconv.Itoa(accountid)
	bindBankCard.BankNo = bankNo
	bindBankCard.BankName = BankName
	bindBankCard.BankAccount = BankAccount
	return bindBankCard
}

func (ctx *BindBankCardReq) BindBankCard() error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return err
	}

	c := base.NewClient(bind_bank_uri, string(b), base.ReqPost)
	c.GenerateSign(base.ReqPost)
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
		return errors.New(base.ErrMsg[res.RespCode])
	}

	return nil
}