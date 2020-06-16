package values

import (
	"ddz/game/player"
)

// 保存所有玩家信息
var (
	UserIDUsers = make(map[int]*player.User)
)

// Broadcast 全服广播
func Broadcast(msg interface{}) {
	for _, user := range UserIDUsers {
		user.WriteMsg(msg)
	}
}
