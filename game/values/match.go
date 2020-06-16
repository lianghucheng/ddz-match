package values

// Match 比赛接口
type Match interface {
	SignIn(uid int) error
	SignOut(uid int) error
	CheckStart() // 判断比赛是否开始
	Start()
	SplitTable()             // 分桌逻辑
	RoundOver(roomID string) // 单局结束，获取结果
	End()
	SendMatchDetail(uid int) // 发送比赛详细
	GetRank(uid int)         // 获取排名情况
}
