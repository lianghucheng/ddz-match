package utils

import (
	"bytes"
	"ddz/conf"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/szxby/tools/log"
)

const (
	key = "7inrmpd5DSQTfDxnAnOH"
)

// PostToAgentServer 向代理后台发送数据
func PostToAgentServer(send interface{}, path string) error {
	log.Debug("post to agent:%v", send)
	params, err := json.Marshal(send)
	if err != nil {
		log.Error("http post call err:%v", err)
		return err
	}
	sign := CalculateMD5(string(params))
	data := map[string]interface{}{"Data": string(params), "Sign": sign}
	reqStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", conf.GetCfgLeafSrv().AgentServer+path, bytes.NewBuffer(reqStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("http post call err:%v", err)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body%v:", string(body))

	// 验证返回参数
	ret := map[string]interface{}{}
	json.Unmarshal(body, &ret)
	if ret["code"] == nil {
		log.Error("call game fail :%v", ret)
		return err
	}
	code, ok := ret["code"].(float64)
	if !ok || code != 0 {
		log.Error("call game fail :%v", ret)
		retMsg := "操作失败，请重试！"
		if ret["desc"] != nil {
			if msg, ok := ret["desc"].(string); ok {
				retMsg = msg
			}
		}
		return errors.New(retMsg)
	}
	return nil
}
