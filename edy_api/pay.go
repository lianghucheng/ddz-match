package edy_api

import (
	"crypto/md5"
	"ddz/config"
	"ddz/game/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/szxby/tools/log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	AppToken  = "fddda32b4cb543babbf78a4ba955c05d"
	AppSecret = "51b793ef1b7e49cf8060e9c083cf17e5"
)

type CommonResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateSign(param string) string {
	shopMerchant := db.ReadShopMerchant()
	merType := strconv.Itoa(shopMerchant.MerchantType)
	cfgPay := config.GetCfgPay()[merType]
	m := md5.New()
	log.Debug("*************%v", cfgPay.AppToken+"&"+param+"&"+cfgPay.AppSecret)
	m.Write([]byte(cfgPay.AppToken + "&" + param + "&" + cfgPay.AppSecret))

	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}
func GenerateSignTemp(param string) string {
	cfgPay := config.GetCfgPay()["1"]
	m := md5.New()
	log.Debug("*************%v", cfgPay.AppToken+"&"+param+"&"+cfgPay.AppSecret)
	m.Write([]byte(cfgPay.AppToken + "&" + param + "&" + cfgPay.AppSecret))

	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

func GetOutTradeNo() string {
	return time.Now().Format("0102150405") + fmt.Sprintf("%05d", rand.Intn(100000))
}

const EdyBackCall = "/edy/pay-bc"

type EdyPayNotifyReq struct {
	AppID       int    `json:"appID"`       //是	应用ID，需要向平台索取
	OpenOrderID string `json:"openOrderID"` //是	开发者订单ID，用户通知及后期对账
	OpenExtend  string `json:"openExtend"`  //是	开发者自定义内容，通知时原文返回
	OrderID     string `json:"orderID"`     //是	订单ID
	OrderTime   string `json:"orderTime"`   //是	下单时间，格式 yyyy-MM-dd HH:mm:ss
	Amount      int    `json:"amount"`      //是	支付金额，百分制，即CNY 1为100
	PayTime     string `json:"payTime"`     //是	支付时间，格式 yyyy-MM-dd HH:mm:ss
	PayType     int    `json:"payType"`     //是	支付类型, 如微信、支付宝等，详细定义见附件
	Ts          int64  `json:"ts"`          //是	通知时的时间戳
	Sign        string `json:"sign"`        //是	签名，详见 4.1 签名计算方法
}

type EdyPayNotifyResp struct {
	OrderResult string `json:"orderResult"` //是	处理结果，success为成功
	OrderAmount string `json:"orderAmount"` //是	处理金额，百分制，100=1元
	OrderTime   string `json:"orderTime"`   //是	处理时间
	Ts          int64  `json:"ts"`          //是	处理完成时的时间戳
	Sign        string `json:"sign"`        //是	签名，此处签名需要开发者按照返回值来计算
}

func GetUrlKeyValStr(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	rt := ""
	cnt := 0
	for k, v := range m {
		if k != "sign" && v != "" {
			if cnt == 0 {
				cnt++
				if data, ok := v.(float64); ok {
					rt += fmt.Sprintf("%v=%v", k, int64(data))
				} else {
					rt += fmt.Sprintf("%v=%v", k, v)
				}
			} else {
				if data, ok := v.(float64); ok {
					rt += fmt.Sprintf("&%v=%v", k, int64(data))
				} else {
					rt += fmt.Sprintf("&%v=%v", k, v)
				}
			}
		}
	}
	strs := strings.Split(rt, "&")
	sort.Strings(strs)
	rt = strings.Join(strs, "&")
	return rt, nil
}
