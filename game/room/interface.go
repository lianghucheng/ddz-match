package room

import (
	. "ddz/game/player"
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
