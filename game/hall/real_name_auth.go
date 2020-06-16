package hall

import (
	"ddz/game"
	"ddz/game/player"
	"ddz/msg"
	"github.com/name5566/leaf/log"
)

func RealNameAuth(user *player.User, msg *msg.C2S_RealNameAuth) {
	realNameAuth(user, msg.IDCardNo, msg.RealName, nil)
}

func realNameAuth(user *player.User, idCardNo, realName string, callRealNameAPI func(idCardNo, realName string) error) {
	if callRealNameAPI == nil {
		user.WriteMsg(&msg.S2C_RealNameAuth{
			Error: msg.ErrRealNameAuthFail,
		})
		return
	}
	if err := callRealNameAPI(idCardNo, realName); err != nil {
		log.Error(err.Error())
		user.WriteMsg(&msg.S2C_RealNameAuth{
			Error: msg.ErrRealNameAuthFail,
		})
		return
	}
	user.BaseData.UserData.RealName = realName
	user.BaseData.UserData.IDCardNo = idCardNo
	game.GetSkeleton().Go(func() {
		player.SaveUserData(user.BaseData.UserData)
	}, func() {
		user.WriteMsg(&msg.S2C_RealNameAuth{
			Error: msg.ErrRealNameAuthSuccess,
		})
		//将实名的数据发送给客户端
	})
}
