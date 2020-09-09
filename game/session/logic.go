package session

import (
	"ddz/conf"
	"ddz/game"
	. "ddz/game/db"
	"ddz/game/hall"
	"ddz/game/http"
	"ddz/game/match"
	. "ddz/game/match"
	. "ddz/game/player"
	. "ddz/game/room"
	"ddz/game/server"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

var skeleton = game.GetSkeleton()

func tokenLogin(user *User, token string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"token": token, "expireat": bson.M{"$gt": time.Now().Unix()}}).One(userData)
		if err != nil {
			log.Debug("find token %v error: %v", token, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_TokenInvalid})
			user.Close()
		}
	}, func() {
		if userData == nil || user.State == UserLogout {
			return
		}
		// 系统维护
		if values.CheckRestart() {
			if !server.CheckWhite(userData.AccountID) {
				user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_ServerRestart, Info: values.DefaultRestartConfig})
				user.Close()
				return
			}
		}

		// 开启白名单模式，白名单可入
		if !server.CheckWhite(userData.AccountID) {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_SystemOff})
			user.Close()
			return
		}
		if userData.Role == RoleBlack {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_RoleBlack})
			user.Close()
			return
		}
		ip := strings.Split(user.RemoteAddr().String(), ":")[0]
		if oldUser, ok := UserIDUsers[userData.UserID]; ok {
			log.Debug("userID: %v 已经登录 %v %v", userData.UserID, oldUser.BaseData.UserData.LoginIP, ip)
			if ip == oldUser.BaseData.UserData.LoginIP {
				oldUser.Close()
			} else {
				user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_IPChanged})
				user.Close()
				return
			}
			user.BaseData = oldUser.BaseData
			userData = oldUser.BaseData.UserData
		}
		UserIDUsers[userData.UserID] = user
		user.BaseData.UserData = userData
		onLogin(user, false, false)
		log.Debug("userID: %v Token登录, 在线人数: %v", userData.UserID, len(UserIDUsers))
	})
}

func usernamePasswordLogin(user *User, account string, password string) {
	userData := new(UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err := db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)
	if err == mgo.ErrNotFound {
		//	firstLogin = true
		//	err = userData.InitValue(0)
		//	userData.Username = account
		//	userData.Headimgurl = DefaultAvatar
		//	if err != nil {
		//		userData = nil
		//		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
		//		user.Close()
		//		return
		//	}
		userData = nil
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Usrn_Nil})
		user.Close()
		return
	}
	log.Debug("*************%+v", userData)
	// firstLogin = userData.FirstLogin
	// if !userData.FirstLogin {
	// 	userData.FirstLogin = !userData.FirstLogin
	//	go func() { SaveUserData(userData) }()
	// }

	if err != nil { //&& err != mgo.ErrNotFound {
		userData = nil
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
		user.Close()
		return
	}
	// 系统维护
	if values.CheckRestart() {
		if !server.CheckWhite(userData.AccountID) {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_ServerRestart, Info: values.DefaultRestartConfig})
			user.Close()
			return
		}
	}
	// 开启白名单模式，白名单可入
	if !server.CheckWhite(userData.AccountID) {
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_SystemOff})
		user.Close()
		return
	}
	if userData.Role == RoleBlack {
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_RoleBlack})
		user.Close()
		return
	}

	// if len(password) < 8 || len(password) > 15 {
	// 	userData = nil
	// 	user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Pass_Length})
	// 	user.Close()
	// 	return
	// }
	if userData.Password != password {
		userData = nil
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Pwd_Error})
		user.Close()
		return
	}

	anotherLogin := false
	log.Debug("player %v login", userData.UserID)
	if oldUser, ok := UserIDUsers[userData.UserID]; ok {
		oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
		oldUser.Close()
		log.Debug("userID: %v 重复登录", userData.UserID)
		if oldUser == user {
			return
		}
		user.BaseData = oldUser.BaseData
		userData = oldUser.BaseData.UserData
	}
	UserIDUsers[userData.UserID] = user
	user.BaseData.UserData = userData
	onLogin(user, userData.FirstLogin, anotherLogin)
}

func logout(user *User) {
	if user.HeartbeatTimer != nil {
		user.HeartbeatTimer.Stop()
	}
	if user.BaseData == nil {
		return
	}
	if existUser, ok := UserIDUsers[user.BaseData.UserData.UserID]; ok {
		if existUser == user {
			log.Debug("userID: %v 登出", user.BaseData.UserData.UserID)
			delete(UserIDUsers, user.BaseData.UserData.UserID)
			user.BaseData.UserData.Online = false
			SaveUserData(user.BaseData.UserData)
		}
	}
	Broadcast(&msg.S2C_OnlineUserNum{
		Num: CalcOnlineCnt(UserIDUsers) + match.GetFakePlayersCount(),
	})
}

func onLogin(user *User, firstLogin bool, anotherLogin bool) {
	log.Debug("***********玩家卡: ")
	log.Debug("***********玩家卡: ")
	log.Debug("***********玩家卡: ")
	log.Debug("***********玩家卡: ")
	log.Debug("***********玩家卡: ")
	user.BaseData.UserData.LoginIP = strings.Split(user.RemoteAddr().String(), ":")[0]
	user.BaseData.UserData.Token = utils.GetToken(32)
	user.BaseData.UserData.ExpireAt = time.Now().Add(2 * time.Hour).Unix()

	user.BaseData.UserData.Online = true
	//if firstLogin {
	//	user.BaseData.UserData.Nickname = "用户" + strconv.Itoa(user.BaseData.UserData.AccountID)
	//	user.BaseData.UserData.Coupon += 5
	//	SaveUserData(user.BaseData.UserData)
	//} else {
	user.RefreshData()
	user.BaseData.UserData.Logintime = time.Now().Unix()
	UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"token": user.BaseData.UserData.Token,
		"online": user.BaseData.UserData.Online, "logintime": time.Now().Unix()}})
	//}
	autoHeartbeat(user)
	bankCard := new(hall.BankCard)
	bankCard.Userid = user.UID()
	bankCard.Read()
	tail := ""
	log.Debug("***********玩家银行卡: %v", bankCard.BankCardNo)
	if bankCard.BankCardNo != "" {
		tail = bankCard.BankCardNo[len(bankCard.BankCardNo)-4:]
	}
	hall.SendDailySignItems(user)
	user.WriteMsg(&msg.S2C_Login{
		AccountID:    user.BaseData.UserData.AccountID,
		Nickname:     user.BaseData.UserData.Nickname,
		Headimgurl:   user.BaseData.UserData.Headimgurl,
		Sex:          user.BaseData.UserData.Sex,
		Role:         user.BaseData.UserData.Role,
		Token:        user.BaseData.UserData.Token,
		AnotherLogin: anotherLogin,
		FirstLogin:   firstLogin,
		//SignIcon:          conf.GetCfgHall().SignIcon,
		//NewWelfareIcon:    conf.GetCfgHall().NewWelfareIcon,
		//FirstRechargeIcon: conf.GetCfgHall().FirstRechargeIcon,
		//ShareIcon:         conf.GetCfgHall().ShareIcon,
		Customer: msg.Customer{
			WeChat:   "wkxjingjipingtai",
			Email:    "jingjikefu@qq.com",
			QQ:       "2721372487",
			QQGroup:  "914746797",
			PhoneNum: "13292757620",
		},
		RealName:       user.RealName(),
		PhoneNum:       user.PhoneNum(),
		BankName:       bankCard.BankName,
		BankCardNoTail: tail,
		SetNickName:    user.GetUserData().SetNickNameCount > 0,
		IsNewUser:      !user.GetUserData().NotNewDailySign,
	})

	// hall.UpdateUserCoupon(user, 0, "")
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
	hall.UpdateUserAfterTaxAward(user)
	hall.SendMail(user)
	hall.SendFirstRecharge(user)
	// hall.SendRaceInfo(user.BaseData.UserData.UserID)
	hall.SendAwardInfo(user)
	hall.SendPriceMenu(user, hall.SendSingle)
	hall.SendPayAccount(user, hall.SendSingle)
	Broadcast(&msg.S2C_OnlineUserNum{
		Num: CalcOnlineCnt(UserIDUsers) + match.GetFakePlayersCount(),
	})
	// 重新为每个赛事存储的玩家副本赋值
	if signs, ok := UserIDSign[user.BaseData.UserData.UserID]; ok {
		for _, v := range signs {
			if _, ok := MatchList[v.SonMatchID]; ok {
				MatchList[v.SonMatchID].AllPlayers[user.BaseData.UserData.UserID] = user
			}
		}
	}
	if s, ok := UserIDMatch[user.BaseData.UserData.UserID]; ok {
		// for uid, p := range s.AllPlayers {
		// 	if p.BaseData.UserData.UserID == user.BaseData.UserData.UserID {

		// 		s.AllPlayers[user.BaseData.UserData.UserID] = user
		// 		break
		// 	}
		// }
		p := s.AllPlayers[user.BaseData.UserData.UserID]
		if user.BaseData.MatchPlayer == nil {
			user.BaseData.MatchPlayer = p.BaseData.MatchPlayer
		}
		s.AllPlayers[user.BaseData.UserData.UserID] = user
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		user.WriteMsg(&msg.S2C_MatchPrepare{})
		skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandloadMatchPrepare)*time.Millisecond, func() {
			r.Enter(user)
		})
	}
}

func AccountLogin(user *User, account string, code string) {
	codeRedis, err := http.GetCaptchaCache(account)
	if err != nil {
		if err == redis.ErrNil {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Code_Error})

			return
		} else {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})

			return
		}
	}
	if code != codeRedis {
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Code_Valid})

		return
	}
	firstLogin := false
	userData := new(UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err = db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)
	if err == mgo.ErrNotFound {
		firstLogin = true
		err = userData.InitValue(0)
		userData.Username = account
		userData.Headimgurl = DefaultAvatar
		if err != nil {
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
			user.Close()
			return
		}
	}
	if err != nil && err != mgo.ErrNotFound {
		userData = nil
		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
		user.Close()
		return
	}
	anotherLogin := false
	if oldUser, ok := UserIDUsers[userData.UserID]; ok {
		oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
		oldUser.Close()
		log.Debug("userID: %v 重复登录", userData.UserID)
		if oldUser == user {
			return
		}
		user.BaseData = oldUser.BaseData
		userData = oldUser.BaseData.UserData
	}
	UserIDUsers[userData.UserID] = user
	user.BaseData.UserData = userData
	onLogin(user, firstLogin, anotherLogin)
}
