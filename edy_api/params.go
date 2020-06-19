package edy_api

import (
	"strconv"
)

type IDBindReq struct {
	CpID 	string `json:"cp_id"`//厂商ID，必填
	PlayerID string`json:"player_id"`//厂商端用户ID，必填
	RealName string`json:"real_name"`//选手名字，必填
	PlayerIDNumber string `json:"player_id_number"` //玩家身份证号码，必填
	PlayerMobile string `json:"player_mobile"`//玩家手机号码，非必填
}

type IDBindResp struct {
	RespCode    string`json:"resp_code"` //编码，Y
	RespMsg     string`json:"resp_msg"` //信息，Y
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

func (ctx *IDBindReq)ToStr() string {
	str := `cp_id=`+ctx.CpID+
		`&player_id=`+ ctx.PlayerID+
		"&real_name="+ctx.RealName+
		"&player_id_number="+ctx.PlayerIDNumber
	if ctx.PlayerMobile != "" {
		str += "&player_mobile=" + ctx.PlayerMobile
	}
	return str
}