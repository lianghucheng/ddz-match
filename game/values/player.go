package values

// GameData 玩家的一些游戏数据
type GameData struct {
	UID       int        `bson:"uid"`
	AccountID int        `bson:"accountid"`
	MatchData *MatchData `bson:"matchdata"`
}

type FlowData struct {
	ID           int `bson:"_id"`
	Userid       int
	Accountid    int
	ChangeAmount float64
	FlowType     int
	MatchType    string
	MatchID      string
	Status       int
	CreatedAt    int64
	FlowIDs      []int
	Realname     string
	TakenFee     float64
	AtferTaxFee  float64
	Desc         string
	PassStatus   int //1是已通过，0是未通过
}