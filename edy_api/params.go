package edy_api

import (
	"strconv"
)

type IDBindReq struct {
	CpID           string `json:"cp_id"`            //厂商ID，必填
	PlayerID       string `json:"player_id"`        //厂商端用户ID，必填
	RealName       string `json:"real_name"`        //选手名字，必填
	PlayerIDNumber string `json:"player_id_number"` //玩家身份证号码，必填
	PlayerMobile   string `json:"player_mobile"`    //玩家手机号码，非必填
}

type IDBindResp struct {
	RespCode string `json:"resp_code"` //编码，Y
	RespMsg  string `json:"resp_msg"`  //信息，Y
}

func NewIDBindReq(accountid int, idCardNo, realName, phoneNum string) *IDBindReq {
	idBind := new(IDBindReq)
	idBind.CpID = cp_id
	idBind.PlayerID = strconv.Itoa(accountid)
	idBind.RealName = realName
	idBind.PlayerIDNumber = idCardNo
	idBind.PlayerMobile = phoneNum
	return idBind
}

func (ctx *IDBindReq) ToStr() string {
	str := `cp_id=` + ctx.CpID +
		`&player_id=` + ctx.PlayerID +
		"&real_name=" + ctx.RealName +
		"&player_id_number=" + ctx.PlayerIDNumber
	if ctx.PlayerMobile != "" {
		str += "&player_mobile=" + ctx.PlayerMobile //todo:可选项是否要去除
	}
	return str
}

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
	bindBankCard.CpID = cp_id
	bindBankCard.PlayerID = strconv.Itoa(accountid)
	bindBankCard.BankNo = bankNo
	bindBankCard.BankName = BankName
	bindBankCard.BankAccount = BankAccount
	return bindBankCard
}
