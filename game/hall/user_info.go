package hall

import (
	"ddz/config"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"github.com/name5566/leaf/log"
	"time"

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

func AddCoupon(user *player.User, count int64) {
	user.BaseData.UserData.Coupon += count
	user.WriteMsg(&msg.S2C_GetCoupon{
		Error: msg.S2C_GetCouponSuccess,
	})
	UpdateUserCoupon(user, count, user.BaseData.UserData.Coupon-count, user.BaseData.UserData.Coupon, db.NormalOpt, db.Charge)
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
		UpdateUserCoupon(user, 5, ud.Coupon-5, ud.Coupon, db.NormalOpt, db.InitPlayer)
	})
}

func SendPriceMenu(user *player.User) {
	user.WriteMsg(&msg.S2C_PriceMenu{
		PriceItems: config.GetPriceMenu(),
	})
}

// todo：暂时PropType当作PropID的作用
type KnapsackProp struct {
	ID        int `bson:"_id"`
	Accountid int
	PropID    int
	Name      string
	Num       int
	IsAdd     bool
	IsUse     bool
	Expiredat int64
	Desc      string
	Createdat int64
}

func (ctx *KnapsackProp) ReadByAidPid() {
	db.Read("knapsackprop", ctx, bson.M{"accountid": ctx.Accountid, "propid": ctx.PropID})
}

func (ctx *KnapsackProp) ReadAllByAid() *[]KnapsackProp {
	knapsackProps := new([]KnapsackProp)
	db.ReadAll("knapsackprop", knapsackProps, bson.M{"accountid": ctx.Accountid})
	return knapsackProps
}

func (ctx *KnapsackProp) Save() {
	db.Save("knapsackprop", ctx, bson.M{"_id": ctx.ID})
}

func SendKnapsack(user *player.User) {
	knapsack := new(KnapsackProp)
	knapsack.Accountid = user.AcountID()
	knapsacks := knapsack.ReadAllByAid()
	user.WriteMsg(&msg.S2C_Knapsack{
		Props: knapsack2Msg(knapsacks),
	})
}

func knapsack2Msg(knapsacks *[]KnapsackProp) *[]msg.KnapsackProp {
	kps := new([]msg.KnapsackProp)
	for _, knapsack := range *knapsacks {
		temp := new(msg.KnapsackProp)
		temp.PropID = knapsack.PropID
		temp.Name = knapsack.Name
		temp.Num = knapsack.Num
		temp.IsUse = knapsack.IsUse
		temp.Expiredat = knapsack.Expiredat
		temp.Desc = knapsack.Desc
		temp.Createdat = temp.Createdat
		*kps = append(*kps, *temp)
	}
	return kps
}

func UseProp(user *player.User, m *msg.C2S_UseProp) {
	ud := user.GetUserData()
	if m.PropID == config.PropIDCouponFrag {
		item, ok := config.PropList[m.PropID]
		if !ok {
			log.Error("no exist prop. ")
			return
		}
		if item.IsUse || item.IsAdd {
			knapsackProp := new(KnapsackProp)
			knapsackProp.Accountid = user.AcountID()
			knapsackProp.PropID = m.PropID
			knapsackProp.ReadByAidPid()
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
				SendKnapsack(user)
				UpdateUserCoupon(user, int64(m.Amount), before, ud.Coupon, db.FragChangeOpt, db.CouponFrag)
				user.WriteMsg(&msg.S2C_UseProp{
					Error:  values.SuccessS2C_UseProp,
					ErrMsg: values.ErrMsg[values.SuccessS2C_UseProp],
				})
			})
		}
	}
}

func AddPropAmount(propid int, accountid int, amount int) {
	knapsackProp := new(KnapsackProp)
	prop, ok := config.PropList[propid]
	if !ok {
		log.Error("没有这个道具配置")
		return
	}
	if prop.IsAdd {
		knapsackProp.Accountid = accountid
		knapsackProp.PropID = propid
		knapsackProp.ReadByAidPid()
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
		knapsackProp.PropID = propid
		knapsackProp.Num += int(amount)
		knapsackProp.Save()
	}
}
