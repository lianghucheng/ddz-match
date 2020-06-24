package hall

import (
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"github.com/name5566/leaf/log"
)

type Realname struct {
	UserID   int
	RealName string
	IDCardNo string
	PhoneNum string
}

func RealNameAuth(user *player.User, m *msg.C2S_RealNameAuth) {
	rn := new(Realname)
	rn.IDCardNo = m.IDCardNo
	rn.RealName = m.RealName
	rn.UserID = user.UID()
	rn.realNameAuth(user, edy_api.RealAuthApi)
}

func (ctx *Realname) realNameAuth(user *player.User, api func(accountid int, idCardNo, realName, phoneNum string) error) {
	if api == nil {
		UpdateRealName(user, msg.ErrRealNameAuthFail,"认证失败")
		return
	}
	if user.RealName() != "" {
		UpdateRealName(user, msg.ErrRealNameAuthAlready,"重复认证")
		return
	}
	var err error
	game.GetSkeleton().Go(func() {
		err = api(ctx.UserID, ctx.IDCardNo, ctx.RealName, ctx.PhoneNum)
	}, func() {
		if err != nil {
			log.Error(err.Error())
			UpdateRealName(user, msg.ErrRealNameAuthBusiness, err.Error())
			return
		}
		ud := user.GetUserData()
		ud.RealName = ctx.RealName
		ud.IDCardNo = ctx.IDCardNo
		player.SaveUserData(ud)
		UpdateRealName(user, msg.ErrRealNameAuthSuccess, "认证成功")
	})
}
