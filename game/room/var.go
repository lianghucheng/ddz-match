package room

// 房间状态
const (
	roomIdle = iota // 0 空闲
	roomGame        // 1 游戏中
)

var (
	roomNumbers []int
	roomCounter = 0
	UserIDRooms = make(map[int]Game)
)

type Room struct {
	State                   int
	LoginIPs                map[string]bool
	PositionUserIDs         map[int]int // key: 座位号, value: userID
	Number                  string
	Desc                    string
	StartTimestamp          int64 // 开始时间
	EachRoundStartTimestamp int64 // 每一局开始时间
	EndTimestamp            int64 // 结束时间
	Game                    Game
}
