package session

import (
	"ddz/conf"
	"ddz/game"
	. "ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/match"
	. "ddz/game/player"
	. "ddz/game/room"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"strconv"
	"strings"
	"time"

	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
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

func usernamePasswordLogin(user *User, account string, code string) {
	//codeRedis, err := GetCaptchaCache(account)
	//if err != nil {
	//	if err == redis.ErrNil {
	//		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Code_Error})
	//
	//		return
	//	} else {
	//		user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
	//
	//		return
	//	}
	//}
	//if code != codeRedis {
	//	user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_Code_Valid})
	//
	//	return
	//}
	firstLogin := false
	userData := new(UserData)
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	// load userData
	err := db.DB(DB).C("users").Find(bson.M{"username": account}).One(userData)
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
}

func onLogin(user *User, firstLogin bool, anotherLogin bool) {
	user.BaseData.UserData.LoginIP = strings.Split(user.RemoteAddr().String(), ":")[0]
	user.BaseData.UserData.Token = utils.GetToken(32)
	user.BaseData.UserData.ExpireAt = time.Now().Add(2 * time.Hour).Unix()

	user.BaseData.UserData.Online = true
	if firstLogin {
		user.BaseData.UserData.Nickname = "用户" + strconv.Itoa(user.BaseData.UserData.AccountID)
		SaveUserData(user.BaseData.UserData)
	} else {
		UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"token": user.BaseData.UserData.Token, "online": user.BaseData.UserData.Online}})
	}
	autoHeartbeat(user)
	user.WriteMsg(&msg.S2C_Login{
		AccountID:         user.BaseData.UserData.AccountID,
		Nickname:          user.BaseData.UserData.Nickname,
		Headimgurl:        user.BaseData.UserData.Headimgurl,
		Sex:               user.BaseData.UserData.Sex,
		Role:              user.BaseData.UserData.Role,
		Token:             user.BaseData.UserData.Token,
		AnotherLogin:      anotherLogin,
		FirstLogin:        firstLogin,
		SignIcon:          conf.GetCfgHall().SignIcon,
		NewWelfareIcon:    conf.GetCfgHall().NewWelfareIcon,
		FirstRechargeIcon: conf.GetCfgHall().FirstRechargeIcon,
		ShareIcon:         conf.GetCfgHall().ShareIcon,
	})

	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.BaseData.UserData.Coupon,
	})
	user.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
		AfterTaxAward: user.BaseData.UserData.Fee,
	})
	hall.SendDailySignItems(user)
	hall.SendFirstRecharge(user)
	hall.SendRaceInfo(user.BaseData.UserData.UserID)
	if s, ok := UserIDMatch[user.BaseData.UserData.UserID]; ok {
		for _, p := range s.AllPlayers {
			if p.BaseData.UserData.UserID == user.BaseData.UserData.UserID {
				s.AllPlayers[user.BaseData.UserData.UserID] = user
				break
			}
		}
	}
	if r, ok := UserIDRooms[user.BaseData.UserData.UserID]; ok {
		user.WriteMsg(&msg.S2C_MatchPrepare{})
		skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandloadMatchPrepare)*time.Millisecond, func() {
			r.Enter(user)
		})
	}
}
