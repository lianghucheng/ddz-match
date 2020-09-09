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

func lengthOfLongestSubstring(s string) int {
	mp := make(map[byte]int)
	i := 0
	max := 0
	for j:=0;j< len(s);j++ {
		if _, ok := mp[s[j]]; ok {
			i = maxval(i, mp[s[j]] + 1)
		}
		mp[s[j]] = j
		max = maxval(max, j - i + 1)
	}
	return max
}

func maxval(a, b int) int{
	if a > b {

		return a
	} else {
		return b
	}
}
