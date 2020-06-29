package http

import (
	. "ddz/game/db"
	"ddz/msg"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	captchaTpl           = "#code#=%s" // 验证码模板
	SUCCESS              = 0
	SYSTEM_ERROR         = 1000
	CAPTCHA_EXPIRE       = 1001
	CAPTCHA_WRONG        = 1002
	INONSISTENT_PASSWORD = 1003
	INVALIDPARAMETER     = 1004
	PHONENUMBER_INVALID  = 1005
	OUTOFSMS             = 1006
	CAPTCHA_SEND_FAIL    = 1007
	ACCOUNTREGISTERD     = 1008
	ACCOUNT_INVALID      = 1009
	PASSWORD_INVALID     = 1010
	FORMAT_FAIL 		 = 1011 //数据格式错误

)

var success = NewError(SUCCESS, "成功")
var systemError = NewError(SYSTEM_ERROR, "系统错误")

const (
	JUHEURL = "http://v.juhe.cn/sms/send"
)

type Error struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewError(errCode int64, errMsg string) *Error {
	return &Error{
		errCode,
		errMsg,
	}
}

func strbyte(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

const (
	CaptchaCacheSeconds = 60 * 2
)

type Register struct {
	Account  string
	Password string
	Code     string
}

func SetCaptchaCache(account string, captcha string) error {
	return Send("SET", "captcha:"+account, captcha, "EX", CaptchaCacheSeconds)
}

func GetCaptchaCache(account string) (captcha string, err error) {
	captcha, err = redis.String(Do("GET", "captcha:"+account))
	return
}
func DelCaptchaCache(account string) error {
	return Send("DEL", "captcha:"+account)
}
func (req *Register) InvalidParameter() bool {
	if req.Account == "" || req.Password == "" || req.Code == "" {
		return true
	}
	return false
}

type SmSJUHEResult struct {
	ErrorCode int32  `json:"error_code"` // 0代表发送成功
	Reason    string `json:"reason"`
	Result    Result `json:"result"`
}

type Result struct {
	Count int    `json:"count"`
	Fee   int    `json:"fee"`
	Sid   string `json:"sid"`
}

func (result *SmSJUHEResult) Success() bool {
	return result.ErrorCode == 0
}

// 单条短信发送,智能匹配短信模板
// apikey 成功注册后登录云片官网,进入后台可查看
// text 需要使用已审核通过的模板或者默认模板
// mobile 接收的手机号,仅支持单号码发送
func JuSend(key string, tplId string, tplValue string, mobile string) (result *SmSJUHEResult, err error) {
	data := url.Values{}
	data.Add("key", key)
	data.Add("tpl_id", tplId)
	data.Add("tpl_value", tplValue)
	data.Add("mobile", mobile)
	respBody, err := PostForm(JUHEURL, data)
	log.Println("************:", string(respBody))
	if err != nil {
		return
	}
	result = &SmSJUHEResult{}
	err = json.Unmarshal(respBody, result)
	return
}

func PostForm(url string, data url.Values) ([]byte, error) {
	response, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http PostForm error : url=%v , statusCode=%v", url, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

type JuHeSmsLog struct {
	Id         string `bson:"_id"`
	ReturnCode int32
	Phone      string
	Captcha    string
	SendTime   int64
	Ip         string
}

func NewJuHeSmsLog(juHeResult *SmSJUHEResult, captcha string, ip string, phone string) *JuHeSmsLog {
	log := &JuHeSmsLog{}
	log.Id = juHeResult.Result.Sid
	log.ReturnCode = juHeResult.ErrorCode
	log.Phone = phone
	log.Captcha = captcha
	log.Ip = ip
	log.SendTime = time.Now().Unix()
	return log
}

func CheckSms(account, code string) int {
	codeRedis, err := GetCaptchaCache(account)
	if err != nil {
		if err == redis.ErrNil {
			return msg.S2C_Close_Code_Error
		} else {
			return msg.S2C_Close_InnerError
		}
	}
	if code != codeRedis {
		return msg.S2C_Close_Code_Valid
	}

	return 0
}
