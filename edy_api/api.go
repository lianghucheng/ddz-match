package edy_api

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	uri = "/player/identity_number/bind"
)

func (ctx *IDBindReq)IDCardBind() error {
	b, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	rt, err := Post(uri, string(b))
	if err != nil {
		return err
	}
	fmt.Println(string(rt))
	res := new(IDBindResp)
	if err := json.Unmarshal(rt, res); err != nil {
		return err
	}

	if res.RespCode != "0" {
		return errors.New("请求接口失败")
	}

	return nil
}