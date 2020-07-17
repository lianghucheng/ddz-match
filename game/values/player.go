package values

// GameData 玩家的一些游戏数据
type GameData struct {
	UID       int        `bson:"uid"`
	AccountID int        `bson:"accountid"`
	MatchData *MatchData `bson:"matchdata"`
}
