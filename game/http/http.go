package http

import (
	"ddz/conf"
	"ddz/game"
	. "ddz/game/db"
	"ddz/game/hall"
	"ddz/msg"
	"ddz/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
	//_ "net/http/pprof"
)

func init() {
	go startHTTPServer()
}

func startHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/code", handleCode)
	mux.HandleFunc("/pushmail", hall.HandlePushMail)
	mux.HandleFunc("/temppay", HandleTempPay)

	err := http.ListenAndServe(conf.GetCfgLeafSrv().HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleCode(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	account := req.FormValue("account")
	if !utils.PhoneRegexp(account) {
		errMsg := NewError(PHONENUMBER_INVALID, "号码不合法")
		w.Write(strbyte(errMsg))
		return
	}
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	nowTime := time.Now()
	start := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local).Unix()
	end := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 24, 0, 0, 0, time.Local).Unix()
	n, err := se.DB(DB).C("juhesmslog").Find(bson.M{"phone": account, "sendtime": bson.M{"$gt": start, "$lt": end}}).Count()
	log.Debug("err:%v", err)
	if err != nil {
		w.Write(strbyte(systemError))
		return
	}
	if n == 6 {
		//errMsg := NewError(OUTOFSMS, "超出短信发送限制")
		//w.Write(strbyte(errMsg))
		//return
	}
	code := utils.RandomNumber(6)
	tplValue := fmt.Sprintf(captchaTpl, code)
	//result, err := SingleSend("b3cbbc5586f0314533a96a52ea3c06dc", text, account)
	juHeResult, err := JuSend(conf.GetCfgJuHeSms().AppKey, conf.GetCfgJuHeSms().RegisterTemplate, tplValue, account)
	log.Debug("%v:", juHeResult)
	if err != nil {
		log.Debug("captcha error, SingleSend error, err=%s,phone=%s", err.Error(), account)
		errMsg := NewError(CAPTCHA_SEND_FAIL, "短信发送失败")
		w.Write(strbyte(errMsg))
		return

	}
	if !juHeResult.Success() {
		log.Debug("captcha error, yunpian.SingleSend error, result.Code=%v,result.Msg=%s,phone=%s", juHeResult.ErrorCode, juHeResult.Reason, account)
		errMsg := NewError(CAPTCHA_SEND_FAIL, "短信发送失败")
		w.Write(strbyte(errMsg))
		return
	}
	ip := strings.Split(req.RemoteAddr, ":")[0]
	logJuHesms := NewJuHeSmsLog(juHeResult, code, ip, account)
	writeJuHeSmsLog(logJuHesms)
	err = SetCaptchaCache(account, code)
	if err != nil {
		log.Debug("captcha error, SetCaptchaCache error, err=%s,phone=%d,captcha=%s", err.Error(), account, code)
		w.Write(strbyte(systemError))
		return
	}
	log.Debug("captcha send success,phone=%s,captcha=%s", account, code)
	w.Write(strbyte(success))
	return
}

func HandleTempPay(w http.ResponseWriter, r *http.Request) {
	secret := r.FormValue("secret")
	aid := r.FormValue("aid")
	fee := r.FormValue("fee")

	f, _ := strconv.Atoi(fee)
	a, _ := strconv.Atoi(aid)
	if secret != "123456" {
		w.Write([]byte("1"))
		return
	}
	w.Write([]byte("0"))
	game.GetSkeleton().ChanRPCServer.Go("TempPayOK", &msg.RPC_TempPayOK{
		TotalFee: f,
		AccountID:a,
	})
}