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
	bind_bank_uri = "/player/bank_acc/bind"
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
	idBind.CpID = base.CpID
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

func (ctx *IDBindReq) IdCardBind() error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return err
	}

	c := base.NewClient(bind_id_uri, string(b), base.ReqPost)
	c.GenerateSign(base.ReqPost)
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
		return errors.New(base.ErrMsg[res.RespCode])
	}

	return nil
}
