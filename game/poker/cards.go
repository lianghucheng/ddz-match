package poker

import (
	"fmt"
	"math"
)

const (
	_           = iota
	DiamondCard // 方块
	ClubCard    // 梅花
	HeartCard   // 红桃
	SpadeCard   // 黑桃
	JokerCard   // 王
)

// 游戏结果
const (
	ResultLose = iota // 0 失败
	ResultWin         // 1 胜利
)

var (
	Diamonds = []int{0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48}  // 方块3到方块A
	Clubs    = []int{1, 5, 9, 13, 17, 21, 25, 29, 33, 37, 41, 45, 49}  // 梅花3到梅花A
	Hearts   = []int{2, 6, 10, 14, 18, 22, 26, 30, 34, 38, 42, 46, 50} // 红桃3到红桃A
	Spades   = []int{3, 7, 11, 15, 19, 23, 27, 31, 35, 39, 43, 47, 51} // 黑桃3到黑桃A
	Jokers   = []int{52, 53}                                           // 小王、大王

	CardType = []int{
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, DiamondCard, ClubCard, HeartCard, SpadeCard,
		DiamondCard, ClubCard, HeartCard, SpadeCard, JokerCard, JokerCard,
	}
	CardString = []string{
		"方块3", "梅花3", "红桃3", "黑桃3", "方块4", "梅花4", "红桃4", "黑桃4",
		"方块5", "梅花5", "红桃5", "黑桃5", "方块6", "梅花6", "红桃6", "黑桃6",
		"方块7", "梅花7", "红桃7", "黑桃7", "方块8", "梅花8", "红桃8", "黑桃8",
		"方块9", "梅花9", "红桃9", "黑桃9", "方块10", "梅花10", "红桃10", "黑桃10",
		"方块J", "梅花J", "红桃J", "黑桃J", "方块Q", "梅花Q", "红桃Q", "黑桃Q",
		"方块K", "梅花K", "红桃K", "黑桃K", "方块A", "梅花A", "红桃A", "黑桃A",
		"方块2", "梅花2", "红桃2", "黑桃2", "小王", "大王",
	}
)

func CardValue(card int) int {
	switch card {
	case 52:
		return 16
	case 53:
		return 17
	}
	return int(math.Floor(float64(card/4)) + 3)
}

func MeldKey(meld []int) string {
	return fmt.Sprintf("%d:%d", meld[0], len(meld))
}

func CountCardValue(cards []int, card int) int {
	count, value := 0, CardValue(card)
	for _, v := range cards {
		if CardValue(v) == value {
			count++
		}
	}
	return count
}

// cards 从大到小
func GetCardsByValue(cards []int, value int) []int {
	a := []int{}
	for _, card := range cards {
		value2 := CardValue(card)
		if value == value2 {
			a = append(a, card)
		} else if value > value2 {
			break
		}
	}
	return a
}

func GetCardsMap(cards []int, m map[int]bool) map[int]bool {
	for _, card := range cards {
		m[card] = true
	}
	return m
}

func GetMeldsCardsMap(melds [][]int, m map[int]bool) map[int]bool {
	for _, meld := range melds {
		m = GetCardsMap(meld, m)
	}
	return m
}

func GetCardValueMap(cards []int) map[int]bool {
	m := make(map[int]bool)
	for _, card := range cards {
		m[CardValue(card)] = true
	}
	return m
}

func RemoveCardByValue(cards []int, valueMap map[int]bool) []int {
	a := []int{}
	for _, card := range cards {
		if !valueMap[CardValue(card)] {
			a = append(a, card)
		}
	}
	return a
}

func GetMeldByValue(melds [][]int, value int) []int {
	for _, meld := range melds {
		valueMap := GetCardValueMap(meld)
		if valueMap[value] {
			return meld
		}
	}
	return []int{}
}

func ToCardsString(cards []int) []string {
	var s []string
	for _, v := range cards {
		s = append(s, CardString[v])
	}
	return s
}

func ToMeldsString(melds [][]int) [][]string {
	s := [][]string{}
	for _, v := range melds {
		s = append(s, ToCardsString(v))
	}
	return s
}

// cards 从大到小
func Sequence(cards []int) bool {
	value := CardValue(cards[0])
	cardsLen := len(cards)
	for i := 1; i < cardsLen; i++ {
		value2 := CardValue(cards[i])
		if value-value2 == 1 {
			value = value2
		} else {
			return false
		}
	}
	return true
}

// 全是对子且连续
func AllPair(cards []int) (bool, bool) {
	cardsLen := len(cards)
	if cardsLen == 0 || cardsLen%2 > 0 {
		return false, false
	}
	allPair, temp := true, []int{}
	for i := 0; i < cardsLen/2; i++ {
		if CountCardValue(cards[i*2:i*2+2], cards[i*2]) == 2 {
			temp = append(temp, cards[i*2])
		} else {
			allPair = false
			break
		}
	}
	if allPair {
		return true, Sequence(temp)
	}
	return false, false
}

// 全是三张且连续
func AllTrio(cards []int) (bool, bool) {
	cardsLen := len(cards)
	if cardsLen == 0 || cardsLen%3 > 0 {
		return false, false
	}
	allTrio, temp := true, []int{}
	for i := 0; i < cardsLen/3; i++ {
		if CountCardValue(cards[i*3:i*3+3], cards[i*3]) == 3 {
			temp = append(temp, cards[i*3])
		} else {
			allTrio = false
			break
		}
	}
	if allTrio {
		return true, Sequence(temp)
	}
	return false, false
}
