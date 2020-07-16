package edy_api

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/szxby/tools/log"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	payHost   = "https://api.test.boai1986.cn"
	appID = 100001
	appToken = "fddda32b4cb543babbf78a4ba955c05d"
	appSecret = "51b793ef1b7e49cf8060e9c083cf17e5"
)

type PayCommon struct {
	VersionCode   int    //是	根据应用发布版本号
	VersionName   string //是	根据应用发布版本名
	Channel       string //否	渠道，最终发行渠道，默认default
	Os            string //是	系统 [android, ios, H5]
	OsVersionName string //是	系统版本名称，如android中的 4.4.0，H5为js版本
	OsVersionCode string //是	系统版本号，如Android的API版本，H5为js版本
	Mac           string //是	MAC地址，全大写，无分隔符，H5使用fingerprint模拟
	DeviceID      string //是	设备ID，H5使用fingerprint模拟
	DeviceModel   string //是	设备型号
}

type CommonResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type CreateOrderReq struct {
	PayCommon
	AppID         int    //是	应用ID，需要向平台索取
	AppToken      string //是	应用授权码，需要向平台索取
	Amount        int    //是	支付金额，百分制，即CNY 1为100
	PayType       int    //是	支付类型, 如微信、支付宝等，详细定义见附件
	Subject       string //是	订单主题
	Description   string //是	订单描述
	Country       string //否	国家码，依据iso3166，默认CN
	Currency      string //否	货币码，依据iso4217，默认CNY
	OpenOrderID   string //否	开发者订单ID，用户通知及后期对账
	OpenExtend    string //否	开发者自定义内容，通知时原文返回
	OpenNotifyURL string //否	开发者回调地址（如果为空或不是URL表示不需要回调，直接充值）
}

type CreateOrderResp struct {
	Result int // 下单状态
	//    1000:订单已创建，等待终端处理
	//    1001:订单成功，针对服务端即可完成支付的情况，如卡密、IPTV
	//    1002:订单失败，针对服务端即可完成支付的情况，如卡密、IPTV
	OrderID  string // 订单ID，后续用于查询订单状态必须
	AppID    int    // 下单应用ID
	PayType  int    // 支付类型，类型定义见附件
	Amount   int    // 支付金额，单位分
	Currency string // 支付金额，单位分
	PayInfo  string // 支付信息，对应不同支付方式返回的内容不同
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateSign(param string) string {
	m := md5.New()
	log.Debug("*************%v",appToken+ "&" + param + "&" + appSecret)
	m.Write([]byte(appToken+ "&" + param + "&" + appSecret))

	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

func GetOutTradeNo() string {
	return time.Now().Format("0102150405") + fmt.Sprintf("%05d", rand.Intn(100000))
}

const EdyBackCall = "/edy/pay-bc"
const CreateOrderUrl = "/payment/create"

func (ctx *CreateOrderReq) Request() (*CreateOrderResp, error) {
	createOrderByte, err := json.Marshal(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req, err := http.NewRequest("POST", payHost+CreateOrderUrl, bytes.NewBuffer(createOrderByte))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debug(string(respByte))
	rt := new(CreateOrderResp)
	return rt, nil
}

type EdyPayNotifyReq struct {
	AppID       int    `json:"appID"`       //是	应用ID，需要向平台索取
	OpenOrderID string `json:"openOrderID"` //是	开发者订单ID，用户通知及后期对账
	OpenExtend  string `json:"openExtend"`  //是	开发者自定义内容，通知时原文返回
	OrderID     string `json:"orderID"`     //是	订单ID
	OrderTime   string `json:"orderTime"`   //是	下单时间，格式 yyyy-MM-dd HH:mm:ss
	Amount      int    `json:"amount"`      //是	支付金额，百分制，即CNY 1为100
	PayTime     string `json:"payTime"`     //是	支付时间，格式 yyyy-MM-dd HH:mm:ss
	PayType     int    `json:"payType"`     //是	支付类型, 如微信、支付宝等，详细定义见附件
	Ts          int64    `json:"ts"`          //是	通知时的时间戳
	Sign        string `json:"sign"`        //是	签名，详见 4.1 签名计算方法
}

type EdyPayNotifyResp struct {
	OrderResult string    `json:"orderResult"` //是	处理结果，success为成功
	OrderAmount string `json:"orderAmount"` //是	处理金额，百分制，100=1元
	OrderTime   string `json:"orderTime"`   //是	处理时间
	Ts          int64    `json:"ts"`          //是	处理完成时的时间戳
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
		if k != "sign" && v != ""{
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
