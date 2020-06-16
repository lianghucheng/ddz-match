package room

import (
	"ddz/utils"
	"fmt"
)

func init() {
	for i := 0; i < 1000000; i++ {
		roomNumbers = append(roomNumbers, i)
	}
	roomNumbers = utils.Shuffle(roomNumbers)
}

func GetRoomNumber() string {
	roomNumber := fmt.Sprintf("%06d", roomNumbers[roomCounter])
	roomCounter = (roomCounter + 1) % 1000000
	return roomNumber
}

func InitRoom() *Room {
	room := new(Room)
	room.LoginIPs = make(map[string]bool)
	room.PositionUserIDs = make(map[int]int)
	room.Number = GetRoomNumber()
	room.State = roomIdle
	return room
}
