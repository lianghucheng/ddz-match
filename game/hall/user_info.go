package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/rpc"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"time"

	"github.com/szxby/tools/log"

	"gopkg.in/mgo.v2/bson"
)

func SetNickname(user *player.User, nickname string) {
	if len(nickname) < 9 || len(nickname) > 18 {
		user.WriteMsg(&msg.S2C_UpdateNickName{
			Error:  msg.S2C_SetNickName_Length,
			ErrMsg: "请输入长度为3-6的汉字",
		})
		return
	}
	ud := user.GetUserData()
	if ud.SetNickNameCount >= 1 {
		user.WriteMsg(&msg.S2C_UpdateNickName{
			Error:  msg.S2C_SetNickName_More,
			ErrMsg: "只能修改一次昵称",
		})
		return
	}
	ud.Nickname = nickname
	ud.SetNickNameCount++
	player.UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"nickname": nickname}})
	user.WriteMsg(&msg.S2C_UpdateNickName{
		Error:    0,
		NickName: nickname,
		ErrMsg:   "修改成功",
	})
}

// AddCoupon 点券变动
func AddCoupon(uid, accountID int, amount int64, opt int, way, matchID string) {
	user, ok := player.UserIDUsers[uid]
	var before, after int64

	// 玩家在线
	if ok {
		user.BaseData.UserData.Coupon += amount
		// user.WriteMsg(&msg.S2C_GetCoupon{
		// 	Error: msg.S2C_GetCouponSuccess,
		// })
		user.WriteMsg(&msg.S2C_UpdateUserCoupon{
			Coupon: user.Coupon(),
		})
		after = user.BaseData.UserData.Coupon
		before = after - amount
	} else {
		s := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(s)
		one := player.UserData{}
		err := s.DB(db.DB).C("users").Find(bson.M{"accountid": accountID}).One(&one)
		if err != nil {
			before = one.Coupon
			after = before + amount
		}
	}
	if amount != 0 {
		game.GetSkeleton().Go(
			func() {
				data := db.ItemLog{
					UID:        accountID,
					Item:       values.Coupon,
					Way:        way,
					Amount:     amount,
					Before:     before,
					After:      after,
					OptType:    opt,
					CreateTime: time.Now().Unix(),
					MatchID:    matchID,
				}
				db.InsertItemLog(data)
				player.UpdateUserData(uid, bson.M{"$inc": bson.M{"coupon": amount}})
			}, nil)
	}
}

// AddFee 奖金变动
func AddFee(uid, accountID int, amount float64, opt int, way, matchID string) {
	user, ok := player.UserIDUsers[uid]
	var before, after int64
	changeAmount := FeeAmount(uid)
	// 玩家在线
	if ok {
		user.GetUserData().Fee = changeAmount
		// game.GetSkeleton().Go(func() {
		// 	player.SaveUserData(user.GetUserData())
		// }, nil)
		UpdateUserAfterTaxAward(user)
	}
	after = int64(changeAmount * 100)
	before = after - int64(amount*100)
	if amount != 0 {
		game.GetSkeleton().Go(
			func() {
				data := db.ItemLog{
					UID:        accountID,
					Item:       "奖金",
					Way:        way,
					Amount:     int64(amount * 100),
					Before:     before,
					After:      after,
					OptType:    opt,
					CreateTime: time.Now().Unix(),
					MatchID:    matchID,
				}
				db.InsertItemLog(data)
				player.UpdateUserData(uid, bson.M{"$set": bson.M{"fee": changeAmount}})
			}, nil)
	}
}

// AddFeeWithTime 奖金变动
func AddFeeWithTime(uid, accountID int, amount float64, opt int, way, matchID string, timestamp int64) {
	user, ok := player.UserIDUsers[uid]
	var before, after int64
	changeAmount := FeeAmount(uid)
	// 玩家在线
	if ok {
		user.GetUserData().Fee = changeAmount
		// game.GetSkeleton().Go(func() {
		// 	player.SaveUserData(user.GetUserData())
		// }, nil)
		UpdateUserAfterTaxAward(user)
	}
	after = int64(changeAmount * 100)
	before = after - int64(amount*100)
	if amount != 0 {
		game.GetSkeleton().Go(
			func() {
				data := db.ItemLog{
					UID:        accountID,
					Item:       "奖金",
					Way:        way,
					Amount:     int64(amount * 100),
					Before:     before,
					After:      after,
					OptType:    opt,
					CreateTime: timestamp,
					MatchID:    matchID,
				}
				db.InsertItemLog(data)
				player.UpdateUserData(uid, bson.M{"$set": bson.M{"fee": changeAmount}})
			}, nil)
	}
}

// AddFragment 碎片奖励
func AddFragment(uid, accountID int, amount int64, opt int, way string, matchID string) {
	data := &KnapsackProp{}
	data.Accountid = accountID
	data.PropType = values.PropTypeCouponFrag
	data.ReadByAidPtype()
	before := int64(data.Num)
	data.Num += int(amount)
	after := int64(data.Num)

	AddPropAmount(data.PropType, accountID, int(amount))

	// data.Save()
	// hall.AddPropAmount(config.PropIDCouponFrag, user.BaseData.UserData.AccountID, int(amount))
	db.InsertItemLog(db.ItemLog{
		UID:        accountID,
		Item:       values.Fragment,
		Amount:     amount,
		Way:        way,
		CreateTime: time.Now().Unix(),
		Before:     before,
		After:      after,
		OptType:    opt,
		MatchID:    matchID,
	})
}

// AddRedScore 红分变动
func AddRedScore(uid, accountID int, amount float64, opt int, way string, matchID string) {
	user, ok := player.UserIDUsers[uid]
	var before, after int64
	sportCenter := player.SportCenterData{}

	// 玩家在线
	if ok {
		user.BaseData.UserData.SportCenter.RedScore += amount
		after = int64(user.BaseData.UserData.SportCenter.RedScore * 100)
		sportCenter = user.BaseData.UserData.SportCenter
	} else {
		s := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(s)
		one := player.UserData{}
		err := s.DB(db.DB).C("users").Find(bson.M{"accountid": accountID}).One(&one)
		if err != nil {
			sportCenter = one.SportCenter
			sportCenter.RedScore += amount
			after = int64(sportCenter.RedScore * 100)
		}
	}
	before = after - int64(amount*100)
	if sportCenter.RedScore != 0 {
		game.GetSkeleton().Go(
			func() {
				db.InsertItemLog(db.ItemLog{
					UID:        user.BaseData.UserData.AccountID,
					Item:       values.RedScore,
					Amount:     int64(amount * 100),
					Way:        way,
					CreateTime: time.Now().Unix(),
					Before:     before,
					After:      after,
					OptType:    opt,
					MatchID:    matchID,
				})
				player.UpdateUserData(uid, bson.M{"$set": bson.M{"sportcenter": sportCenter}})
			}, nil)
	}
}

func UpdateUserAfterTaxAward(user *player.User) {
	changeAmount := FeeAmount(user.UID())
	user.GetUserData().Fee = changeAmount
	game.GetSkeleton().Go(func() {
		player.SaveUserData(user.GetUserData())
	}, nil)
	user.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
		AfterTaxAward: utils.Decimal(changeAmount),
	})
}

func ChangePassword(user *player.User, m *msg.C2S_ChangePassword) {
	ud := user.GetUserData()
	if ud.Password != m.OldPassword {
		user.WriteMsg(&msg.S2C_ChangePassword{
			Error: msg.ErrChangePasswordOldNo,
		})
		return
	}

	ud.Password = m.NewPassword

	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
		user.WriteMsg(&msg.S2C_ChangePassword{
			Error: msg.ErrChangePasswordSuccess,
		})
	}, nil)
}

func TakenFirstCoupon(user *player.User) {
	ud := user.GetUserData()
	if ud.FirstLogin == false {
		user.WriteMsg(&msg.S2C_TakenFirstCoupon{
			Error: msg.ErrS2CTakenFirstCouponFail,
		})
		return
	}
	ud.FirstLogin = false
	ud.Coupon += 5
	game.GetSkeleton().Go(func() {
		player.SaveUserData(ud)
	}, func() {
		user.WriteMsg(&msg.S2C_TakenFirstCoupon{})
		UpdateUserCoupon(user, 5, ud.Coupon-5, ud.Coupon, db.ActivityOpt, db.InitPlayer)
	})
}

// todo：暂时PropType当作PropID的作用
type KnapsackProp struct {
	ID        int `bson:"_id"`
	Accountid int
	PropType  int
	Name      string
	Num       int
	IsAdd     bool
	IsUse     bool
	Expiredat int64
	Desc      string
	Createdat int64
}

func (ctx *KnapsackProp) ReadByAidPtype() {
	db.Read("knapsackprop", ctx, bson.M{"accountid": ctx.Accountid, "proptype": ctx.PropType})
}

func (ctx *KnapsackProp) ReadAllByAid() *[]KnapsackProp {
	knapsackProps := new([]KnapsackProp)
	db.ReadAll("knapsackprop", knapsackProps, bson.M{"accountid": ctx.Accountid, "num": bson.M{"$ne": 0}})
	return knapsackProps
}

func (ctx *KnapsackProp) Save() {
	db.Save("knapsackprop", ctx, bson.M{"_id": ctx.ID})
}

func SendKnapsack(user *player.User) {
	knapsack := new(KnapsackProp)
	knapsack.Accountid = user.AcountID()
	knapsacks := knapsack.ReadAllByAid()
	log.Debug("%+v", *knapsack2Msg(knapsacks))
	user.WriteMsg(&msg.S2C_Knapsack{
		Props: *knapsack2Msg(knapsacks),
	})
}

func knapsack2Msg(knapsacks *[]KnapsackProp) *[]msg.KnapsackProp {
	kps := new([]msg.KnapsackProp)
	for _, knapsack := range *knapsacks {
		temp := new(msg.KnapsackProp)
		temp.PropID = knapsack.PropType
		temp.Name = config.GetPropBaseConfig(knapsack.PropType).Name
		temp.Num = knapsack.Num
		temp.IsUse = knapsack.IsUse
		temp.Expiredat = knapsack.Expiredat
		temp.Desc = knapsack.Desc
		temp.Createdat = temp.Createdat
		temp.Imgurl = config.GetPropBaseConfig(knapsack.PropType).ImgUrl
		*kps = append(*kps, *temp)
	}
	return kps
}

func UseProp(user *player.User, m *msg.C2S_UseProp) {
	ud := user.GetUserData()
	if m.PropID == values.PropTypeCouponFrag {
		item, ok := config.PropList[m.PropID]
		if !ok {
			log.Error("no exist prop. ")
			return
		}
		if item.IsUse || item.IsAdd {
			if m.Amount == 0 {
				user.WriteMsg(&msg.S2C_UseProp{
					Error:  values.ErrS2C_UsePropCouponFragLack,
					ErrMsg: values.ErrMsg[values.ErrS2C_UsePropCouponFragLack],
				})
				return
			}
			knapsackProp := new(KnapsackProp)
			knapsackProp.Accountid = user.AcountID()
			knapsackProp.PropType = m.PropID
			knapsackProp.ReadByAidPtype()
			if knapsackProp.Num < m.Amount*20 {
				user.WriteMsg(&msg.S2C_UseProp{
					Error:  values.ErrS2C_UsePropCouponFragLack,
					ErrMsg: values.ErrMsg[values.ErrS2C_UsePropCouponFragLack],
				})
				return
			}
			knapsackProp.Num -= m.Amount * 20
			before := ud.Coupon
			ud.Coupon += int64(m.Amount)
			game.GetSkeleton().Go(func() {
				player.SaveUserData(ud)
				knapsackProp.Save()
			}, func() {
				UpdateUserCoupon(user, int64(m.Amount), before, ud.Coupon, db.FragChangeOpt, db.CouponFrag)
				cpItem := config.PropList[values.PropTypeCoupon]
				user.WriteMsg(&msg.S2C_UseProp{
					Error:  values.SuccessS2C_UseProp,
					ErrMsg: values.ErrMsg[values.SuccessS2C_UseProp],
					Amount: m.Amount,
					Name:   config.GetPropBaseConfig(cpItem.ID).Name,
					PropID: cpItem.ID,
					Imgurl: config.GetPropBaseConfig(cpItem.ID).ImgUrl,
				})
			})
		}
	}
}

func AddPropAmount(propType int, accountid int, amount int) {
	log.Debug("道具数量变动，propid：%v， accountid： %v, amount:%v. ", propType, accountid, amount)
	knapsackProp := new(KnapsackProp)
	prop, ok := config.PropList[propType]
	if !ok {
		log.Error("没有这个道具配置")
		return
	}
	if prop.IsAdd {
		knapsackProp.Accountid = accountid
		knapsackProp.PropType = propType
		knapsackProp.ReadByAidPtype()
		if knapsackProp.Createdat == 0 {
			knapsackProp.ID, _ = db.MongoDBNextSeq("knapsackprop")
			knapsackProp.Createdat = time.Now().Unix()
			knapsackProp.Name = prop.Name
			knapsackProp.IsAdd = prop.IsAdd
			knapsackProp.IsUse = prop.IsUse
			knapsackProp.Expiredat = prop.Expiredat
			knapsackProp.Desc = prop.Desc
		}
		knapsackProp.Num += int(amount)
		knapsackProp.Save()
	} else {
		// todo：需求暂时不用
		knapsackProp.ID, _ = db.MongoDBNextSeq("knapsackprop")
		knapsackProp.Createdat = time.Now().Unix()
		knapsackProp.Name = prop.Name
		knapsackProp.IsAdd = prop.IsAdd
		knapsackProp.IsUse = prop.IsUse
		knapsackProp.Expiredat = prop.Expiredat
		knapsackProp.Desc = prop.Desc
		knapsackProp.Accountid = accountid
		knapsackProp.PropType = propType
		knapsackProp.Num += int(amount)
		knapsackProp.Save()
	}
}

//todo:目前只有点券
func SendGoods(userID, amount int) {
	ud := player.ReadUserDataByID(userID)
	if user, ok := player.UserIDUsers[userID]; ok {
		user.GetUserData().Coupon += int64(amount)
		go func() {
			player.SaveUserData(user.GetUserData())
		}()
		user.WriteMsg(&msg.S2C_GetCoupon{
			Error: msg.ErrPaySuccess,
		})
		UpdateUserCoupon(user, int64(amount), user.GetUserData().Coupon-int64(amount), user.GetUserData().Coupon, db.ChargeOpt, db.Charge)
	} else {
		ud.Coupon += int64(amount)
		go func() {
			player.SaveUserData(ud)
		}()
	}
}

func AddSundries(propType int, ud *player.UserData, amount float64, opt int, way, matchID string) {
	switch propType {
	case values.PropTypeAward:
		WriteFlowData(ud.UserID, amount, FlowTypeGift, "", "", []int{}, nil)
		AddFee(ud.UserID, ud.AccountID, amount, opt,
			way, matchID)
	case values.PropTypeCoupon:
		AddCoupon(ud.UserID, ud.AccountID, int64(amount), opt,
			way, matchID)
	case values.PropTypeCouponFrag:
		AddFragment(ud.UserID, ud.AccountID, int64(amount), opt,
			way, matchID)
	case values.PropTypeRedScore:
		AddRedScore(ud.UserID, ud.AccountID, amount, opt,
			way, matchID)
	}
}

// DrawDailyWelfare 领取每日福利
func DrawDailyWelfare(u *player.User, dailyType, awardIndex int) {
	var err error
	reply := &rpc.RPCRet{}
	game.GetSkeleton().Go(func() {
		err = rpc.CallActivityServer("DailyWelfareObj.DrawDailyWelfare",
			rpc.RPCDrawDailyWelfare{AccountID: u.BaseData.UserData.AccountID, DailyType: dailyType, AwardIndex: awardIndex}, reply)
	}, func() {
		log.Debug("player %v draw daily welfare", u.BaseData.UserData.AccountID)
		code := reply.Code
		desc := reply.Desc
		name := ""
		propID := 0
		amount := float64(0)
		imgUrl := ""
		defer func() {
			u.WriteMsg(&msg.S2C_DrawDailyWelfareInfo{
				Code: code,
				Desc: desc,
				Name: name,
				PropID :propID,
				Amount:amount,
				ImgUrl:imgUrl,
			})
		}()
		if err != nil {
			code = 1
			desc = "领取失败!"
			if len(reply.Desc) > 0 {
				desc = reply.Desc
			}
			return
		}
		if reply.Data == nil {
			code = 1
			desc = "领取失败!"
			return
		}
		b, ok := reply.Data.([]byte)
		if !ok {
			code = 1
			desc = "领取失败!"
			return
		}
		var data = struct {
			ItemID int
			Amount int
		}{}
		if err := json.Unmarshal(b, &data); err != nil {
			code = 1
			desc = "领取失败!"
			return
		}
		log.Debug("player %v draw daily welfare:%v", u.BaseData.UserData.AccountID, data)
		// ok
		id := values.PropID2Type[data.ItemID]
		name = config.GetPropBaseConfig(id).Name
		propID = id
		amount = float64(data.Amount)
		imgUrl = config.GetPropBaseConfig(id).ImgUrl
		AddSundries(id, u.BaseData.UserData, float64(data.Amount), db.ActivityOpt, db.DailyWelfare, "")
	})
}
