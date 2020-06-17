package session

import (
	"ddz/conf"
	. "ddz/game/player"
	. "ddz/game/room"
	"ddz/msg"
	"time"

	"github.com/name5566/leaf/gate"
	"github.com/szxby/tools/log"
)

// 用户状态
const (
	UserLogin  = iota
	UserLogout // 1
)

type Game interface {
	OnInit(room *Room)
	Enter(user *User) bool
	Exit(userId int)
	SitDown(user *User, pos int)
	StandUp(user *User, pos int)
	GetAllPlayers(user *User)
	Play(command interface{}, userId int)
	StartGame()
	EndGame()
}

var (
// UserIDUsers = make(map[int]*User)
// UserIDRooms = make(map[int]Game)
// UserIDMatch = make(map[int]*ScoreMatch)
)

func newUser(a gate.Agent) *User {
	user := new(User)
	user.Agent = a
	user.State = UserLogin
	user.BaseData = new(BaseData)
	user.BaseData.UserData = new(UserData)
	return user
}

func autoHeartbeat(user *User) {
	if user.HeartbeatStop {
		log.Debug("userID: %v 心跳停止", user.BaseData.UserData.UserID)
		user.Close()
		return
	}
	user.HeartbeatStop = true
	user.WriteMsg(&msg.S2C_Heartbeat{})
	// 服务端发送心跳包间隔120秒
	user.HeartbeatTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().HeartTimeout)*time.Second, func() {
		autoHeartbeat(user)
	})
}

func isRobot(user *User) bool {
	return user.BaseData.UserData.Role == RoleRobot
}
