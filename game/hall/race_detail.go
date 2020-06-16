package hall

import (
	"ddz/game"
	"ddz/msg"
)

func SendRaceInfo(userid int) {
	game.GetSkeleton().ChanRPCServer.Go("SendRaceInfo", &msg.RPC_SendRaceInfo{ID: userid})
}
