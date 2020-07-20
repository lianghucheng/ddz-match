package http

import (
	"ddz/conf"
	"ddz/edy_api"
	"ddz/game"
	. "ddz/game/db"
	"ddz/game/hall"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/findpwd", handleFindPwd)
	mux.HandleFunc("/edyht-add-fee", handleEdyhtAddFee)
	mux.HandleFunc("/update-coupon", handleUpdateCoupon)

	// 后台比赛接口
	mux.HandleFunc("/addMatch", addMatch)
	mux.HandleFunc("/showHall", showHall)
	mux.HandleFunc("/editSort", editSort)
	mux.HandleFunc("/editMatch", editMatch)
	mux.HandleFunc("/optMatch", optMatch)
	mux.HandleFunc("/optUser", optUser)
	mux.HandleFunc("/clearRealInfo", clearRealInfo)

	mux.HandleFunc("/addaward", addAward)
	mux.HandleFunc("/update-headimg", updateHeadImg)
	//电竞二打一支付回调
	mux.HandleFunc(edy_api.EdyBackCall, edyPayBackCall)

	err := http.ListenAndServe(conf.GetCfgLeafSrv().HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleCode(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := req.FormValue("data")
	log.Debug("data   %v", data)
	temp := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &temp)
	if err != nil {
		errMsg := NewError(PHONENUMBER_INVALID, "号码不合法")
		w.Write(strbyte(errMsg))
		return
	}
	account := temp["Account"].(string)
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
	log.Debug("模板号 %v", conf.GetCfgJuHeSms().RegisterTemplate)
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
		TotalFee:  f,
		AccountID: a,
	})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := r.FormValue("data")
	m := new(msg.C2S_Register)
	err := json.Unmarshal([]byte(data), m)
	if err != nil {
		log.Debug("数据格式错误, %v,err:%v", string(data), err)
		errMsg := NewError(FORMAT_FAIL, "数据格式错误")
		w.Write(strbyte(errMsg))
		return
	}
	//todo:没问题之后再加密
	if len(m.Password) < 8 {
		log.Debug("密码不够8位")
		w.Write(strbyte(NewError(PASSWORD_LACK, "密码不能少于8位")))
		return
	}
	account, code, password, shareCode := m.Account, m.Code, m.Password, m.ShareCode
	_ = code
	if len(shareCode) == 0 {
		w.Write(strbyte(NewError(FORMAT_FAIL, "邀请码不能为空!")))
		return
	}

	if status := CheckSms(account, code); status != 0 {
		log.Debug("status:%v", status)
		w.Write(strbyte(NewError(int64(status), "验证码错误")))
		return
	}
	userData := new(player.UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err = db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)

	if err == nil {
		userData = nil
		w.Write(strbyte(NewError(msg.S2C_Close_Usrn_Exist, "用户名已存在")))
		return
	}

	err = userData.InitValue(0)
	if err != nil {
		userData = nil
		w.Write(strbyte(NewError(msg.S2C_Close_InnerError, "注册失败")))
		return
	}
	// 发送代理后台检查代理情况
	if err := utils.PostToAgentServer(struct {
		ShareCode         string
		RegisterAccountID int
	}{
		ShareCode:         shareCode,
		RegisterAccountID: userData.AccountID,
	}); err != nil {
		w.Write(strbyte(NewError(FORMAT_FAIL, err.Error())))
		return
	}

	userData.Username = account
	userData.Password = password
	userData.Headimgurl = player.DefaultAvatar

	player.SaveUserData(userData)
	w.Write(strbyte(NewError(msg.ErrRegisterSuccess, "注册成功")))
}

func handleFindPwd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := r.FormValue("data")
	log.Debug("%v", string(data))
	m := new(msg.C2S_FindPassword)
	err := json.Unmarshal([]byte(data), m)
	if err != nil {
		log.Debug("数据格式错误, %v", data)
		errMsg := NewError(FORMAT_FAIL, "数据格式错误")
		w.Write(strbyte(errMsg))
		return
	}
	account, code, password := m.Account, m.Code, m.Password
	_ = code
	if status := CheckSms(account, code); status != 0 {
		w.Write(strbyte(NewError(int64(status), "验证码错误")))
		return
	}

	userData := new(player.UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err = db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)

	if err != nil {
		userData = nil
		w.Write(strbyte(NewError(msg.S2C_Close_Usrn_Nil, "用户名不存在")))
		return
	}

	userData.Password = password
	player.SaveUserData(userData)
	w.Write(strbyte(NewError(msg.ErrFindPasswordSuccess, "成功")))
	return
}

func handleEdyhtAddFee(w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	m := new(msg.RPC_AddFee)
	if err := json.Unmarshal([]byte(data), m); err != nil {
		log.Error(err.Error())
		return
	}

	game.GetSkeleton().ChanRPCServer.Go("AddFee", m)
}

func addAward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}
	m := new(msg.RPC_AddAward)
	err = json.Unmarshal(b, m)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if m.Secret != "123456" {
		log.Error("非法调用")
		return
	}

	ud := player.ReadUserDataByAid(m.Uid)
	game.GetSkeleton().ChanRPCServer.Go("AddAward", &msg.RPC_AddAward{
		Uid:    ud.UserID,
		Amount: m.Amount,
	})
	w.Write([]byte(fmt.Sprintf(`{"code": 0, "msg": "%v添加奖金记录成功"}`, m.Uid)))
}

func edyPayBackCall(w http.ResponseWriter, r *http.Request) {
	edyPayNotifyReq := new(edy_api.EdyPayNotifyReq)
	//todo:解析到CreateOrderReq
	edyPayNotifyReq.Amount, _ = strconv.Atoi(r.FormValue("amount"))
	edyPayNotifyReq.AppID, _ = strconv.Atoi(r.FormValue("appID"))
	edyPayNotifyReq.OpenExtend = r.FormValue("openExtend")
	edyPayNotifyReq.OpenOrderID = r.FormValue("openOrderID")
	edyPayNotifyReq.OrderID = r.FormValue("orderID")
	edyPayNotifyReq.OrderTime = r.FormValue("orderTime")
	edyPayNotifyReq.PayType, _ = strconv.Atoi(r.FormValue("payType"))
	edyPayNotifyReq.PayTime = r.FormValue("payTime")
	edyPayNotifyReq.Ts, _ = strconv.ParseInt(r.FormValue("ts"), 10, 64)
	edyPayNotifyReq.Sign = r.FormValue("sign")
	log.Debug("【请求参数】%+v", *edyPayNotifyReq)
	param, err := edy_api.GetUrlKeyValStr(edyPayNotifyReq)
	if err != nil {
		log.Error(err.Error())
		return
	}
	sign := edy_api.GenerateSign(param)
	log.Debug("【生成的签名】%v", sign)
	if sign != edyPayNotifyReq.Sign {
		log.Debug("sign error. ")
		return
	}

	//todo:存订单，发货
	order := new(values.EdyOrder)
	Read("edyorder", order, bson.M{"tradeno": edyPayNotifyReq.OpenOrderID, "status": false})
	order.TradeNoReceive = edyPayNotifyReq.OrderID
	order.Status = true
	Save("edyorder", order, bson.M{"_id": order.ID})
	game.GetSkeleton().ChanRPCServer.Go("TempPayOK", &msg.RPC_TempPayOK{
		TotalFee:  int(order.Fee),
		AccountID: order.Accountid,
	})
	log.Debug("【发货成功】")
	edyPayNotifyResp := new(edy_api.EdyPayNotifyResp)
	edyPayNotifyResp.OrderResult = "success"
	edyPayNotifyResp.OrderAmount = fmt.Sprintln(order.Fee)
	ts := time.Now().Unix()
	edyPayNotifyResp.OrderTime = time.Unix(ts, 0).Format("2006-01-02 03:04:05")
	edyPayNotifyResp.Ts = ts
	param2, err := edy_api.GetUrlKeyValStr(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	edyPayNotifyResp.Sign = edy_api.GenerateSign(param2)
	//todo:封装响应数据
	b, err := json.Marshal(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func handleUpdateCoupon(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err.Error())
		return
	}
	m := new(msg.RPC_UpdateCoupon)
	if err := json.Unmarshal(b, m); err != nil {
		log.Error(err.Error())
		return
	}

	if m.Secret != "123456" {
		log.Error("非法调用")
		return
	}
	game.GetSkeleton().ChanRPCServer.Go("UpdateCoupon", m)
}

func updateHeadImg(w http.ResponseWriter, r *http.Request) {
	log.Debug("【更新头像】")
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err.Error())
		return
	}
	m := new(msg.RPC_UpdateHeadImg)
	if err := json.Unmarshal(b, m); err != nil {
		log.Error(err.Error())
		return
	}

	log.Debug("*********%+v", *m)

	if m.Secret != "123456" {
		log.Error("非法调用")
		return
	}
	game.GetSkeleton().ChanRPCServer.Go("UpdateHeadImg", m)
}
