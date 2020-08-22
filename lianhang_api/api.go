package lianhang_api

import (
	"ddz/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/name5566/leaf/log"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
)

type LianHangReq struct {
	Bank     string `json:"bank"`
	Bankcard string `json:"bankcard"`
	City     string `json:"city"`
	Key      string `json:"key"`
	Province string `json:"province"`
}

//type LianHangResp struct {
//	Msg string
//	Success bool
//	Code int
//	"data": {
//	"order_no": "647016226459095040", //订单号
//	"result": {
//	"totalpage": 1, //总页数
//	"totalcount": 1, //总记录数
//	"bank": "招商银行", //输入的银行名称
//	"province": "浙江", //输入的省
//	"city": "杭州", //输入的市
//	"record": [
//{
//	"bank": "招商银行", //总行名称
//	"lname": "招商银行股份有限公司杭州城西支行", //支行名称
//	"lng": "", //经度
//	"province": "浙江省", //省
//	"city": "杭州市", //市
//	"district": "", //所在区
//	"isHead": "0",//总行标识 0-否 1-是
//	"tel": "0571-88911759", //支行电话
//	"id": "9860293b7d4745728d9b4e1991c0f9d9",//无效字段
//	"addr": "杭州市西湖区文一西路170号", //支行地址
//	"bankcode": "308331012167", //联行号
//	"lat": "" //纬度
//}
//],
//	"page": 1, //输入的页数
//	"card": "", //输入的卡号
//	"key": "西湖" //输入的关键字
//}
//}
//}
//

func (ctx *LianHangReq) LianHangApi() (string, error) {
	cfg := config.GetCfgLianHang()
	req, err := http.NewRequest("GET", cfg.Host+cfg.LianHangUrl+"?"+ToUrlStr(ctx), nil)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	req.Header.Set("Authorization", "APPCODE "+cfg.AppCode)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	log.Debug("########   %v", string(b))
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		log.Error(err.Error())
		return "", err
	}
	code, ok := m["code"]
	if !ok || code.(float64) != 200 {
		return "", errors.New("request lian hang error. ")
	}
	log.Debug("the result map %v   %v", m["data"].(map[string]interface{})["result"].(map[string]interface{})["totalcount"], reflect.TypeOf(m["data"].(map[string]interface{})["result"].(map[string]interface{})["totalcount"]))
	totalCount, ok := m["data"].(map[string]interface{})["result"].(map[string]interface{})["totalcount"]
	if !ok {
		return "", errors.New("lian hang api no totalcount. ")
	}
	i_totalCount, _ := strconv.Atoi(fmt.Sprintf("%v", totalCount))
	if i_totalCount != 1 {
		return "", nil
	}

	bankcode, ok := m["data"].(map[string]interface{})["result"].(map[string]interface{})["record"].([]interface{})[0].(map[string]interface{})["bankcode"].(string)
	if !ok {
		return "", errors.New("lian hang aou no bankcode")
	}
	return bankcode, nil
}

func ToUrlStr(m interface{}) string {
	str := ``
	if m == nil {
		return ""
	}
	jsonD, _ := json.Marshal(m)
	data := map[string]interface{}{}
	json.Unmarshal(jsonD, &data)
	for k, v := range data {
		str += fmt.Sprintf("&%v=%v", k, v)
	}

	return str[1:]
}
