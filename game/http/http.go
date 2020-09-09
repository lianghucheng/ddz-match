package http

import (
	"ddz/cache"
	"ddz/conf"
	"ddz/config"
	"ddz/edy_api"
	"ddz/game"
	. "ddz/game/db"
	"ddz/game/hall"
	"ddz/game/pay"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"errors"
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
	mux.HandleFunc("/pushmail", handlePushMail)
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
	mux.HandleFunc("/editWhiteList", editWhiteList)
	mux.HandleFunc("/getOnline", getOnline)
	mux.HandleFunc("/restart", restartServer)
	mux.HandleFunc("/dealIllegalMatch", dealIllegalMatch)

	mux.HandleFunc("/addaward", addAward)
	mux.HandleFunc("/update-headimg", updateHeadImg)
	mux.HandleFunc("/give-coupon-frag", giveCouponFrag)
	//电竞二打一支付回调
	mux.HandleFunc("/edy/pay-bc", edyPayBackCall)
	mux.HandleFunc("/edy/pay-bc-temp", edyPayBackCallTemp)
	mux.HandleFunc("/conf/robot-maxnum", confRobotMaxNum)
	mux.HandleFunc("/add/coupon-frag", addCouponFrag)
	mux.HandleFunc("/notify/payaccount", notifyPayAccount)
	mux.HandleFunc("/notify/pricemenu", notidyPriceMenu)
	mux.HandleFunc("/set/propbaseconfig", setPropBaseConfig)
	mux.HandleFunc("/reload/config", reloadConfig)

	mux.HandleFunc("/faker/create-order", fakerCreateOrder)
	mux.HandleFunc("/faker/valid-order", fakerValidOrder)

	mux.HandleFunc("/pushmail/notify-clean", pushMailNotifyClean)
	mux.HandleFunc("/bind/lianhanghao", handleLianHangHao)

	mux.HandleFunc("/test", handleTest)
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
	log.Debug("真人美女斗地主，远程调用支付", aid, fee)
	f, _ := strconv.Atoi(fee)
	a, _ := strconv.Atoi(aid)
	if secret != "123456" {
		w.Write([]byte("1"))
		return
	}
	w.Write([]byte("0"))
	order := new(values.EdyOrder)
	order.TradeNo = utils.GetOutTradeNo()
	order.Fee = int64(f)
	order.Amount = f / 100
	order.Createdat = time.Now().Unix()
	order.ID, _ = MongoDBNextSeq("edyorder")
	order.Accountid = a
	order.Merchant = values.ZrddzAliPay
	order.Status = true
	order.PayStatus = 1
	Save("edyorder", order, bson.M{"_id": order.ID})
	game.GetSkeleton().ChanRPCServer.Go("NotifyPayOK", &msg.RPC_NotifyPayOK{
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
	if len(m.Password) < 8 || len(m.Password) > 15 {
		log.Debug("密码长度为8～15位")
		w.Write(strbyte(NewError(PASSWORD_LACK, "密码长度为8～15位")))
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
	}, "/bindAgent"); err != nil {
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

	if len(password) < 8 || len(password) > 15 {
		log.Debug("密码长度为8～15位")
		w.Write(strbyte(NewError(PASSWORD_LACK, "密码长度为8～15位")))
		return
	}

	userData := new(player.UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err = db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)

	if err != nil {
		userData = nil
		w.Write(strbyte(NewError(msg.S2C_Close_Usrn_Nil, "用户不存在")))
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

	//存订单
	order := new(values.EdyOrder)
	Read("edyorder", order, bson.M{"tradeno": edyPayNotifyReq.OpenOrderID, "status": false})
	if order.PayStatus != values.PayStatusAction {
		log.Debug("支付失败，不发货")
		return
	}
	order.TradeNoReceive = edyPayNotifyReq.OrderID
	order.Status = true
	order.PayStatus = values.PayStatusSuccess
	Save("edyorder", order, bson.M{"_id": order.ID})
	//通知完成支付，进行发货
	game.GetSkeleton().ChanRPCServer.Go("NotifyPayOK", &msg.RPC_NotifyPayOK{
		TotalFee:  int(order.Fee),
		AccountID: order.Accountid,
		Amount:    order.Amount,
	})
	//通知第三方支付流程已完成，封装响应数据
	edyPayNotifyResp := new(edy_api.EdyPayNotifyResp)
	edyPayNotifyResp.OrderResult = "success"
	edyPayNotifyResp.OrderAmount = fmt.Sprintf("%v", order.Fee)
	ts := time.Now().Unix()
	edyPayNotifyResp.OrderTime = time.Unix(ts, 0).Format("2006-01-02 03:04:05")
	edyPayNotifyResp.Ts = ts
	param2, err := edy_api.GetUrlKeyValStr(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	edyPayNotifyResp.Sign = edy_api.GenerateSign(param2)
	b, err := json.Marshal(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func edyPayBackCallTemp(w http.ResponseWriter, r *http.Request) {
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
	sign := edy_api.GenerateSignTemp(param)
	log.Debug("【生成的签名】%v", sign)
	if sign != edyPayNotifyReq.Sign {
		log.Debug("sign error. ")
		return
	}

	//存订单
	order := new(values.EdyOrder)
	Read("edyorder", order, bson.M{"tradeno": edyPayNotifyReq.OpenOrderID, "status": false})
	if order.PayStatus != values.PayStatusAction {
		log.Debug("支付失败，不发货")
		return
	}
	order.TradeNoReceive = edyPayNotifyReq.OrderID
	order.Status = true
	order.PayStatus = values.PayStatusSuccess
	Save("edyorder", order, bson.M{"_id": order.ID})
	//通知完成支付，进行发货
	game.GetSkeleton().ChanRPCServer.Go("NotifyPayOK", &msg.RPC_NotifyPayOK{
		TotalFee:  int(order.Fee),
		AccountID: order.Accountid,
		Amount:    order.Amount,
	})
	//通知第三方支付流程已完成，封装响应数据
	edyPayNotifyResp := new(edy_api.EdyPayNotifyResp)
	edyPayNotifyResp.OrderResult = "success"
	edyPayNotifyResp.OrderAmount = fmt.Sprintf("%v", order.Fee)
	ts := time.Now().Unix()
	edyPayNotifyResp.OrderTime = time.Unix(ts, 0).Format("2006-01-02 03:04:05")
	edyPayNotifyResp.Ts = ts
	param2, err := edy_api.GetUrlKeyValStr(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	edyPayNotifyResp.Sign = edy_api.GenerateSign(param2)
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

func giveCouponFrag(w http.ResponseWriter, r *http.Request) {
	accountid := r.FormValue("accountid")
	amount := r.FormValue("amount")
	log.Debug("后台发放点券:aid:%v   amount:%v", accountid, amount)
	aid, _ := strconv.Atoi(accountid)
	a, _ := strconv.Atoi(amount)
	propType := values.PropTypeCouponFrag
	prop, ok := config.PropList[propType]
	if !ok {
		log.Error("没有这个道具配置")
		return
	}
	knapsack := new(hall.KnapsackProp)
	knapsack.PropType = propType
	knapsack.Accountid = aid
	knapsack.ReadByAidPtype()
	if knapsack.Createdat == 0 {
		knapsack.ID, _ = MongoDBNextSeq("knapsackprop")
		knapsack.Createdat = time.Now().Unix()
		knapsack.Name = prop.Name
		knapsack.IsAdd = prop.IsAdd
		knapsack.IsUse = prop.IsUse
		knapsack.Expiredat = prop.Expiredat
		knapsack.Desc = prop.Desc
	}
	knapsack.Num += int(a)
	knapsack.Save()
	w.Write([]byte(`1`))
	game.GetSkeleton().ChanRPCServer.Go("SendKnapsack", &msg.RPC_SendKnapsack{
		Aid: aid,
	})
}

func confRobotMaxNum(w http.ResponseWriter, r *http.Request) {
	maxRobotNum := r.FormValue("max_robot_num")
	matchid := r.FormValue("matchid")
	num, _ := strconv.Atoi(maxRobotNum)
	log.Debug("更新赛事机器人最大数量配置，配置前数量：%v", config.GetCfgMatchRobotMaxNums()[matchid])
	config.GetCfgMatchRobotMaxNums()[matchid] = num
	log.Debug("更新赛事机器人最大数量配置，配置后数量：%v", config.GetCfgMatchRobotMaxNums()[matchid])
	//todo: Save
	w.Write([]byte(`更新赛事机器人最大数量配置，配置后数量：` + strconv.Itoa(config.GetCfgMatchRobotMaxNums()[matchid])))
}

func addCouponFrag(w http.ResponseWriter, r *http.Request) {
	data := new(msg.RPC_AddCouponFrag)
	b, errReadIO := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	errJson := json.Unmarshal(b, data)
	if errReadIO != nil || errJson != nil {
		log.Error("errReadIO: %v, errJson: %v. ", errReadIO, errJson)
	}
	if data.Secret != "123456" {
		log.Debug("非法调用")
		return
	}
	game.GetSkeleton().ChanRPCServer.Go("AddCouponFrag", data)
	w.Write([]byte(`1`))
}

func notifyPayAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 10000
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	hall.SendPayAccount(nil, hall.SendBroacast)
}

func notidyPriceMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 10000
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	hall.SendPriceMenu(nil, hall.SendBroacast)
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	log.Debug("the http test")
}

func handlePushMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	data := r.FormValue("data")
	log.Debug("调试发送邮件   %v", data)
	param := new(hall.MailBoxParam)
	if err := json.Unmarshal([]byte(data), param); err != nil {
		log.Error(err.Error())
		code = FORMAT_FAIL
		errmsg = ErrMsg[code]
		return
	}

	param.PushMailBox()
}

func setPropBaseConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	if err := cache.UpdatePropBaseConfig(); err != nil {
		log.Error(err.Error())
		code = UPDATE_PROP_CONF_FAIL
		errmsg = ErrMsg[code]
		return
	}
	return
}

func reloadConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	if err := config.UpdateCfg2(); err != nil {
		code = RELOAD_CONFIG_FAIL
		errmsg = ErrMsg[code] + err.Error()
		return
	}

	if err := cache.UpdatePropBaseConfig(); err != nil {
		code = UPDATE_PROP_CONF_FAIL
		errmsg = ErrMsg[code] + err.Error()
		return
	}

	return
}

func fakerCreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	aid := r.FormValue("aid")
	pid := r.FormValue("pid")
	accountid, _ := strconv.Atoi(aid)
	priceid, _ := strconv.Atoi(pid)
	code, errmsg = pay.FakerCreateOrder(accountid, priceid)

	return
}

func fakerValidOrder(w http.ResponseWriter, r *http.Request) {
	OpenOrderID := r.FormValue("openOrderID")
	log.Debug("【请求参数】%+v", OpenOrderID)

	//存订单
	order := new(values.EdyOrder)
	Read("edyorder", order, bson.M{"tradeno": OpenOrderID, "status": false})
	if order.PayStatus != values.PayStatusAction {
		log.Debug("支付失败，不发货")
		return
	}
	order.TradeNoReceive = "faker"
	order.Status = true
	order.PayStatus = values.PayStatusSuccess
	Save("edyorder", order, bson.M{"_id": order.ID})
	//通知完成支付，进行发货
	game.GetSkeleton().ChanRPCServer.Go("NotifyPayOK", &msg.RPC_NotifyPayOK{
		TotalFee:  int(order.Fee),
		AccountID: order.Accountid,
		Amount:    order.Amount,
	})
	//通知第三方支付流程已完成，封装响应数据
	edyPayNotifyResp := new(edy_api.EdyPayNotifyResp)
	edyPayNotifyResp.OrderResult = "success"
	edyPayNotifyResp.OrderAmount = fmt.Sprintf("%v", order.Fee)
	ts := time.Now().Unix()
	edyPayNotifyResp.OrderTime = time.Unix(ts, 0).Format("2006-01-02 03:04:05")
	edyPayNotifyResp.Ts = ts
	param2, err := edy_api.GetUrlKeyValStr(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	edyPayNotifyResp.Sign = edy_api.GenerateSign(param2)
	b, err := json.Marshal(edyPayNotifyResp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func pushMailNotifyClean(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	uds := new([]player.UserData)
	if err := se.DB(DB).C("users").Find(nil).All(uds); err != nil {
		log.Error(err.Error())
		return
	}
	successCnt := 0
	failCnt := 0
	zeroCnt := 0
	for _, ud := range *uds {
		fee := hall.FeeAmount(ud.UserID)
		if fee <= 0 {
			zeroCnt++
			continue
		}
		targetID := fmt.Sprintf("%v", ud.UserID)
		title := fmt.Sprintf("%v", fee)
		content := fmt.Sprintf("亲爱的竞技二打一选手：\n      很高兴的通知大家，体总赛事已经正式上线，您的提奖金额【%v】元请尽快联系客服进行进行人工提现操作，之后将由体总监管账户进行下发奖金！客服联系方式为：wkxjingjipingtai", fee)
		param := fmt.Sprintf(`{"target_id":%v,"mail_type":1,"mail_service_type":0,"title":"%v","content":"%v","annexes":[],"expire_value":10000}`,
			targetID, title, content)
		resp, err := http.Get("http://111.230.39.198:9084/pushmail?data=" + param)
		if err != nil {
			log.Error(err.Error())
			failCnt++
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			failCnt++
			continue
		}
		rt := &struct {
			Code   int
			Errmsg string
		}{}
		if err := json.Unmarshal(b, rt); err != nil {
			log.Error(err.Error())
			failCnt++
			continue
		}
		if rt.Code != 0 {
			err = errors.New(fmt.Sprintf("request fail. code is %v. ", rt.Code))
			log.Error(err.Error())
			failCnt++
			continue
		}

		successCnt++
	}

	errmsg = fmt.Sprintf("总共扫描了：%v个用户。   成功处理了%v个用户。   失败处理了%v个用户。   有%v个用户可提现奖金为0。", len(*uds), successCnt, failCnt, zeroCnt)
}

func handleLianHangHao(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := 0
	errmsg := "success"
	result := make(map[string]interface{})
	defer func() {
		result["code"] = code
		result["errmsg"] = errmsg
		b, err := json.Marshal(result)
		if err != nil {
			log.Error(err.Error())
			return
		}
		i, err := w.Write(b)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Debug("success size:%v. ", i)
	}()

	accountid := r.FormValue("accountid")
	bankName := r.FormValue("bankName")
	bankCardNo := r.FormValue("bankCardNo")
	province := r.FormValue("province")
	city := r.FormValue("city")
	openingBank := r.FormValue("openingBank")
	openingBankNo := r.FormValue("openingBankNo")

	if accountid == "" || bankName == "" || bankCardNo == "" || province == "" || city == "" || openingBank == "" || openingBankNo == "" {
		code = 1
		errmsg = "参数为空"
	}

	aid, _ := strconv.Atoi(accountid)
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	ud := new(player.UserData)
	if err := se.DB(DB).C("users").Find(bson.M{"accountid": aid}).One(ud); err != nil {
		log.Error(err.Error())
		return
	}

	b := &hall.BankCard{
		Userid:        ud.UserID,
		BankName:      bankName,
		BankCardNo:    bankCardNo,
		Province:      province,
		City:          city,
		OpeningBank:   openingBank,
		OpeningBankNo: openingBankNo,
	}
	if err := edy_api.BandBankCardAPI(aid, b.OpeningBankNo, b.BankName, b.BankCardNo); err != nil {
		code = 1
		errmsg = err.Error()
		return
	}
	ud.BankCardNo = b.BankCardNo
	if _, err := se.DB(DB).C("bankcard").Upsert(bson.M{"userid": b.Userid}, b); err != nil {
		code = 1
		errmsg = err.Error()
		return
	}

	game.GetSkeleton().ChanRPCServer.Go("UpdateBankCardNo", &msg.RPC_UpdateBankCardNo{
		Userid:     ud.UserID,
		BankCardNo: ud.BankCardNo,
	})
}
