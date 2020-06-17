package hall

import (
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"github.com/name5566/leaf/log"
)

type RealName struct {
	IDCardNo   string
	RealName   string
}

func RealNameAuth(user *player.User, msg *msg.C2S_RealNameAuth) {
	realName := new(RealName)
	realName.RealName = msg.RealName
	realName.IDCardNo = msg.IDCardNo
	realName.realNameAuth(user, RealNameAPI)
}

func (ctx *RealName)realNameAuth(user *player.User, callRealNameAPI func(idCardNo, realName string) error) {
	if callRealNameAPI == nil {
		UpdateRealName(user, msg.ErrRealNameAuthFail)
		return
	}
	if user.RealName() != "" {
		UpdateRealName(user, msg.ErrRealNameAuthAlready)
		return
	}
	if err := callRealNameAPI(ctx.IDCardNo, ctx.RealName); err != nil {
		log.Error(err.Error())
		UpdateRealName(user, msg.ErrRealNameAuthFail)
		return
	}
	user.BaseData.UserData.RealName = ctx.RealName
	user.BaseData.UserData.IDCardNo = ctx.IDCardNo
	game.GetSkeleton().Go(func() {
		player.SaveUserData(user.BaseData.UserData)
	}, nil)
	UpdateRealName(user, msg.ErrRealNameAuthSuccess)
}

func RealNameAPI(idCardNo, realName string) error {
	return nil
}