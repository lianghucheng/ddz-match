package player

// Broadcast 全服广播
func Broadcast(msg interface{}) {
	for _, user := range UserIDUsers {
		user.WriteMsg(msg)
	}
}
