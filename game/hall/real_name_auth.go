package hall

import (
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"errors"
	"github.com/name5566/leaf/log"
)

func RealNameAuth(user *player.User, msg *msg.C2S_RealNameAuth) {
	idBind := edy_api.NewIDBindReq(user.AcountID(),msg.IDCardNo, msg.RealName, "")
	if err := realNameAuth(user, idBind); err == nil {
		user.BaseData.UserData.RealName = msg.RealName
		user.BaseData.UserData.IDCardNo = msg.IDCardNo
	}
}

func realNameAuth(user *player.User, idBindApi edy_api.IDBindApi) error {
	if idBindApi == nil {
		UpdateRealName(user, msg.ErrRealNameAuthFail)
		return errors.New("msg.ErrRealNameAuthFail")
	}
	if user.RealName() != "" {
		UpdateRealName(user, msg.ErrRealNameAuthAlready)
		return errors.New("msg.ErrRealNameAuthAlready")
	}
	if err := idBindApi.IDCardBind(); err != nil {
		log.Error(err.Error())
		UpdateRealName(user, msg.ErrRealNameAuthFail)
		return errors.New("msg.ErrRealNameAuthFail")
	}
	game.GetSkeleton().Go(func() {
		player.SaveUserData(user.BaseData.UserData)
	}, nil)
	UpdateRealName(user, msg.ErrRealNameAuthSuccess)
	return nil
}