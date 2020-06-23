package poker

import (
	"ddz/utils"
	"sort"
)

// 出牌动作类型
const (
	_                                = iota
	ActionLandlordDiscardMust        // 1 必须
	ActionLandlordDiscardAlternative // 2 可选择的（可出可不出）
	ActionLandlordDiscardNothing     // 3 要不起
)

// 斗地主牌型
const (
	Error            = iota // 0 牌型错误
	Solo                    // 1 单张
	Pair                    // 2 对子
	Trio                    // 3 三张
	TrioSolo                // 4 三带一
	TrioPair                // 5 三带一对
	SoloChain               // 6 顺子
	PairSisters             // 7 连对
	AirplaneChain           // 8 飞机
	TrioSoloAirplane        // 9 飞机带小翼
	TrioPairChain           // 10 飞机带大翼
	FourDualsolo            // 11 四带两单
	FourDualpair            // 12 四带两对
	Bomb                    // 13 炸弹
	KingBomb                // 14 王炸
)

var (
	LandlordAllCards   = landlordAllCards()
	LandlordAllCards2P = landlordAllCards2P()
)

type LstPoker []LandlordPlayerRoundResult

func (a LstPoker) Len() int { // 重写 Len() 方法
	return len(a)
}

func (a LstPoker) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}

func (a LstPoker) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	lable := a[i].Total > a[j].Total ||
		a[i].Total == a[j].Total && a[i].Last > a[j].Last ||
		a[i].Total == a[j].Total && a[i].Last == a[j].Last && a[i].Wins > a[j].Wins ||
		a[i].Total == a[j].Total && a[i].Last == a[j].Last && a[i].Wins == a[j].Wins && a[i].Time < a[j].Time ||
		a[i].Total == a[j].Total && a[i].Last == a[j].Last && a[i].Wins == a[j].Wins && a[i].Time == a[j].Time && a[i].Sort < a[j].Sort

	return lable

}

// 玩家单局成绩
type LandlordPlayerRoundResult struct {
	Uid int
	// Position int    // 玩家位置
	Nickname string // 昵称(用户绑定的真实姓名)
	Wins     int    // 获胜次数
	Chips    int64  `json:"-"` // 筹码
	Total    int64  // 总得分
	Last     int64  // 尾副牌得分
	Time     int64  // 累计用时(单位毫秒)
	Sort     int    // 报名排序
	// Continue bool   // 是否晋级下一局
}

// LandlordRankData 比赛排行榜信息
type LandlordRankData struct {
	Position int    // 玩家位置
	Nickname string // 昵称(用户绑定的真实姓名)
	Wins     int    // 获胜次数
	Total    int64  // 总得分
	Last     int64  // 尾副牌得分
	Time     int64  // 累计用时(单位毫秒)
	Sort     int    // 报名排序
}

// 斗地主所有的扑克牌
func landlordAllCards() []int {
	cards := append([]int{}, Diamonds...)
	cards = append(cards, Clubs...)
	cards = append(cards, Hearts...)
	cards = append(cards, Spades...)
	cards = append(cards, Jokers...)
	return cards
}

// 二人斗地主所有扑克牌
func landlordAllCards2P() []int {
	cards := append([]int{}, Diamonds[2:]...)
	cards = append(cards, Clubs[2:]...)
	cards = append(cards, Hearts[2:]...)
	cards = append(cards, Spades[2:]...)
	cards = append(cards, Jokers...)
	return cards
}

func ToLandlordTypeString(cardsType int) string {
	switch cardsType {
	case Solo: // 1 单张
		return "单张"
	case Pair: // 2 对子
		return "对子"
	case Trio: // 3 三张
		return "三张"
	case TrioSolo: // 4 三带一
		return "三带一"
	case TrioPair: // 5 三带一对
		return "三带一对"
	case SoloChain: // 6 顺子
		return "顺子"
	case PairSisters: // 7 连对
		return "连对"
	case AirplaneChain: // 8 飞机
		return "飞机"
	case TrioSoloAirplane: // 9 飞机带小翼
		return "飞机带小翼"
	case TrioPairChain: // 10 飞机带大翼
		return "飞机带大翼"
	case FourDualsolo: // 11 四带两单
		return "四带两单"
	case FourDualpair: // 12 四带两对
		return "四带两对"
	case Bomb: // 13 炸弹
		return "炸弹"
	case KingBomb: // 14 王炸
		return "王炸"
	default:
		return "牌型错误"
	}
}

// 与手牌比较大小
func CompareLandlordHands(discards []int, hands []int) bool {
	discardsType := GetLandlordCardsType(discards)
	switch discardsType {
	case KingBomb:
		return true
	case Error:
		return false
	}
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(hands)
	if analyzer.hasKingBomb {
		return false
	}
	if discardsType == Bomb {
		if analyzer.hasBomb && CardValue(discards[0]) < CardValue(analyzer.bombs[0][0]) {
			return false
		}
		return true
	}
	if analyzer.hasBomb {
		return false
	}
	discardsLen, handsLen := len(discards), len(hands)
	if discardsLen > handsLen || discardsType == FourDualsolo || discardsType == FourDualpair {
		return true
	}
	switch discardsType {
	case Solo:
		if CardValue(discards[0]) < CardValue(hands[0]) {
			return false
		}
		return true
	case Pair:
		if analyzer.hasPair && CardValue(discards[0]) < CardValue(analyzer.pairs[0][0]) {
			return false
		}
		return true
	case Trio, TrioSolo:
		if analyzer.hasTrio && CardValue(discards[0]) < CardValue(analyzer.trios[0][0]) {
			return false
		}
		return true
	case TrioPair:
		if analyzer.hasTrio && CardValue(discards[0]) < CardValue(analyzer.trios[0][0]) {
			remain := utils.Remove(hands, analyzer.trios[0])
			for _, card := range remain {
				if CountCardValue(remain, card) > 1 {
					return false
				}
			}
		}
		return true
	case SoloChain:
		if analyzer.hasSoloChain && discardsLen <= len(analyzer.soloChains[0]) && CardValue(discards[0]) < CardValue(analyzer.soloChains[0][0]) {
			return false
		}
		return true
	case PairSisters:
		if analyzer.hasPairSisters && discardsLen <= len(analyzer.pairSisterss[0]) && CardValue(discards[0]) < CardValue(analyzer.pairSisterss[0][0]) {
			return false
		}
		return true
	case AirplaneChain: // 飞机
		if analyzer.hasAirplaneChain && discardsLen <= len(analyzer.airplaneChains[0]) && CardValue(discards[0]) < CardValue(analyzer.airplaneChains[0][0]) {
			return false
		}
		return true
	case TrioSoloAirplane: // 飞机带小翼
		discardAirplaneLen := discardsLen / 4 * 3
		discardAirplane := discards[:discardAirplaneLen]
		if analyzer.hasAirplaneChain && discardAirplaneLen <= len(analyzer.airplaneChains[0]) && CardValue(discardAirplane[0]) < CardValue(analyzer.airplaneChains[0][0]) {
			remain := utils.Remove(hands, analyzer.airplaneChains[0][:discardAirplaneLen])
			if len(remain) >= discardsLen/4 {
				return false
			}
		}
		return true
	case TrioPairChain: // 飞机带大翼
		discardAirplaneLen := discardsLen / 5 * 3
		discardAirplane := discards[:discardAirplaneLen]
		if analyzer.hasAirplaneChain && discardAirplaneLen <= len(analyzer.airplaneChains[0]) && CardValue(discardAirplane[0]) < CardValue(analyzer.airplaneChains[0][0]) {
			remain, countPair := utils.Remove(hands, analyzer.airplaneChains[0][:discardAirplaneLen]), 0
			for _, card := range remain {
				if CountCardValue(remain, card) > 1 {
					countPair++
				}
			}
			if countPair >= discardsLen/5 {
				return false
			}
		}
		return true
	}
	return false
}

// 与出的牌比较大小
func CompareLandlordDiscard(discards []int, preDiscards []int) bool {
	discardsType := GetLandlordCardsType(discards)
	preDiscardsType := GetLandlordCardsType(preDiscards)
	if discardsType == preDiscardsType {
		if len(discards) == len(preDiscards) && CardValue(discards[0]) > CardValue(preDiscards[0]) {
			return true
		}
		return false
	}
	switch discardsType {
	case KingBomb, Bomb:
		return true
	}
	return false
}

func GetLandlordCardsType(cards []int) int {
	cardsLen := len(cards)
	if cardsLen == 1 {
		return Solo // 单张
	}
	if cardsLen == 2 {
		if cards[0] == 53 && cards[1] == 52 {
			return KingBomb // 王炸
		}
		if CardValue(cards[0]) == CardValue(cards[1]) {
			return Pair // 对子
		}
	}
	if cardsLen == 3 {
		if CountCardValue(cards, cards[0]) == 3 {
			return Trio // 三张
		}
	}
	if cardsLen == 4 {
		count1stCardValue := CountCardValue(cards, cards[0])
		if count1stCardValue == 4 {
			return Bomb // 炸弹
		}
		if count1stCardValue == 3 || CountCardValue(cards[1:], cards[1]) == 3 {
			return TrioSolo // 三带一
		}
	}
	if cardsLen == 5 {
		for i := 0; i < 2; i++ {
			if CountCardValue(cards[i*2:i*2+3], cards[i*2]) == 3 {
				remain := utils.Remove(cards, cards[i*2:i*2+3])
				if CardValue(remain[0]) == CardValue(remain[1]) {
					return TrioPair // 三带一对
				}
			}
		}
	}
	if cardsLen >= 5 {
		if CardValue(cards[0]) < 15 && Sequence(cards) { // < 2
			return SoloChain // 顺子
		}
	}
	if cardsLen == 6 {
		for i := 0; i < 3; i++ {
			if CountCardValue(cards[i:i+4], cards[i]) == 4 {
				return FourDualsolo // 四带两单
			}
		}
	}
	if cardsLen >= 6 && cardsLen%2 == 0 {
		if CardValue(cards[0]) < 15 { // < 2
			allPair, seq := AllPair(cards)
			if allPair && seq {
				return PairSisters // 连对
			}
		}
	}
	if cardsLen >= 6 && cardsLen%3 == 0 {
		if CardValue(cards[0]) < 15 { // < 2
			allTrio, seq := AllTrio(cards)
			if allTrio && seq {
				return AirplaneChain // 飞机
			}
		}
	}
	if cardsLen == 8 {
		for i := 0; i < 3; i++ {
			if CountCardValue(cards[i*2:i*2+4], cards[i*2]) == 4 {
				remain := utils.Remove(cards, cards[i*2:i*2+4])
				if allPair, _ := AllPair(remain); allPair && CountCardValue(remain, remain[0]) == 2 {
					return FourDualpair // 四带两对
				}
			}
		}
	}
	if cardsLen >= 8 && cardsLen%4 == 0 {
		for i := 0; i <= cardsLen/4; i++ { // 遍历小翼的个数
			if CardValue(cards[i]) < 15 { // < 2
				allTrio, seq := AllTrio(cards[i : i+cardsLen/4*3])
				if allTrio && seq {
					valueMap := GetCardValueMap(cards[i : i+cardsLen/4*3])
					remain := RemoveCardByValue(cards, valueMap)
					if len(remain) == cardsLen/4 {
						return TrioSoloAirplane // 飞机带小翼
					}
				}
			}
		}
	}
	if cardsLen >= 10 && cardsLen%5 == 0 {
		for i := 0; i <= cardsLen/5; i++ { // 遍历大翼的个数
			if CardValue(cards[i*2]) < 15 { // < 2
				allTrio, seq := AllTrio(cards[i*2 : i*2+cardsLen/5*3])
				if allTrio && seq {
					valueMap := GetCardValueMap(cards[i*2 : i*2+cardsLen/5*3])
					remain := RemoveCardByValue(cards, valueMap)
					allPair, _ := AllPair(remain)
					if allPair && len(remain) == cardsLen/5*2 {
						return TrioPairChain // 飞机带大翼
					}
				}
			}
		}
	}
	return Error
}

// cards 的长度至少为6且是3的倍数
func GetLandlordAirplaneChains(cards []int) [][]int {
	airplaneChains := [][]int{}
	cardsLen := len(cards)
	if cardsLen > 5 && cardsLen%3 == 0 {
		temp := append([]int{}, cards[:3]...)
		for i := 1; i < cardsLen/3; i++ {
			trio := cards[i*3 : i*3+3]
			if Sequence([]int{temp[len(temp)-1], trio[0]}) {
				temp = append(temp, trio...)
			} else {
				if len(temp) > 5 {
					airplaneChains = append(airplaneChains, temp)
				}
				temp = append([]int{}, trio...)
			}
		}
		if len(temp) > 5 {
			airplaneChains = append(airplaneChains, temp)
		}
	}
	return airplaneChains
}

// cards 的长度至少为6且是2的倍数
func GetLandlordPairSisters(cards []int) [][]int {
	pairSisters := [][]int{}
	cardsLen := len(cards)
	if cardsLen > 5 && cardsLen%2 == 0 {
		temp := append([]int{}, cards[:2]...)
		for i := 1; i < cardsLen/2; i++ {
			pair := cards[i*2 : i*2+2]
			if CardValue(temp[len(temp)-1]) == CardValue(pair[0]) {
				continue
			}
			if Sequence([]int{temp[len(temp)-1], pair[0]}) {
				temp = append(temp, pair...)
			} else {
				if len(temp) > 5 {
					pairSisters = append(pairSisters, temp)
				}
				temp = append([]int{}, pair...)
			}
		}
		if len(temp) > 5 {
			pairSisters = append(pairSisters, temp)
		}
	}
	return pairSisters
}

// cards 的长度至少为5
func GetLandlordSoloChains(cards []int) [][]int {
	cardsLen := len(cards)
	if cardsLen < 5 {
		return [][]int{}
	}
	soloChains := [][]int{}
	temp := []int{cards[0]}
	for i := 1; i < cardsLen; i++ {
		solo := cards[i]
		if CardValue(temp[len(temp)-1]) == CardValue(solo) {
			continue
		}
		if Sequence([]int{temp[len(temp)-1], solo}) {
			temp = append(temp, solo)
		} else {
			if len(temp) > 4 {
				soloChains = append(soloChains, temp)
			}
			temp = []int{solo}
		}
	}
	if len(temp) > 4 {
		soloChains = append(soloChains, temp)
	}
	return soloChains
}

func GetDiscardHint(preDiscards []int, hands []int) [][]int {
	preDiscardsType := GetLandlordCardsType(preDiscards)
	switch preDiscardsType {
	case Solo: // 单张
		return GetGreaterThanSolo(CardValue(preDiscards[0]), hands)
	case Pair: // 对子
		return GetGreaterThanPair(CardValue(preDiscards[0]), hands)
	case Trio: // 三张
		return GetGreaterThanTrio(CardValue(preDiscards[0]), hands)
	case TrioSolo: // 三带一
		return GetGreaterThanTrioSolo(CardValue(preDiscards[0]), hands)
	case TrioPair: // 三带一对
		return GetGreaterThanTrioPair(CardValue(preDiscards[0]), hands)
	case SoloChain: // 顺子
		return GetGreaterThanSoloChain(preDiscards, hands)
	case PairSisters: // 连对
		return GetGreaterThanPairSisters(preDiscards, hands)
	case AirplaneChain: // 飞机
		return GetGreaterThanAirplaneChain(preDiscards, hands)
	case TrioSoloAirplane: // 飞机带小翼
		return GetGreaterThanTrioSoloAirplane(preDiscards, hands)
	case TrioPairChain: // 飞机带大翼
		return GetGreaterThanTrioPairChain(preDiscards, hands)
	case FourDualsolo, FourDualpair: // 四带两单、四带两对
		return GetBombs(hands)
	case Bomb: // 炸弹
		return GetGreaterThanBomb(CardValue(preDiscards[0]), hands)
	}
	return [][]int{}
}

func GetBombs(cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	bombs := [][]int{}
	if analyzer.hasBomb {
		bombsLen := len(analyzer.bombs)
		for i := bombsLen - 1; i > -1; i-- {
			bombs = append(bombs, analyzer.bombs[i])
		}
	}
	if analyzer.hasKingBomb {
		bombs = append(bombs, []int{53, 52})
	}
	return bombs
}

// 获取大于单张的牌
func GetGreaterThanSolo(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	m, melds := make(map[int]bool), [][]int{}
	remain := append([]int{}, cards...)
	if analyzer.hasSoloChain {
		for _, soloChain := range analyzer.soloChains {
			remain = utils.Remove(remain, soloChain)
		}
	}
	sort.Ints(remain)
	for _, card := range remain {
		cardValue2 := CardValue(card)
		if m[cardValue2] {
			continue
		}
		m[cardValue2] = true
		if cardValue2 > cardValue {
			melds = append(melds, []int{card})
		}
	}
	melds = ReSortHint(Solo, melds, remain)
	if analyzer.hasSoloChain {
		temp := utils.Remove(cards, remain)
		sort.Ints(temp)
		for _, card := range temp {
			cardValue2 := CardValue(card)
			if m[cardValue2] {
				continue
			}
			m[cardValue2] = true
			if cardValue2 > cardValue {
				melds = append(melds, []int{card})
			}
		}
	}
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于对子的牌
func GetGreaterThanPair(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	if analyzer.hasPair {
		m := make(map[int]bool)
		pairsLen := len(analyzer.pairs)
		for i := pairsLen - 1; i > -1; i-- {
			value := CardValue(analyzer.pairs[i][0])
			if m[value] {
				continue
			}
			m[value] = true
			if value > cardValue {
				melds = append(melds, analyzer.pairs[i])
			}
		}
	}
	melds = ReSortHint(Pair, melds, cards)
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于三张的牌
func GetGreaterThanTrio(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	if analyzer.hasTrio {
		triosLen := len(analyzer.trios)
		for i := triosLen - 1; i > -1; i-- {
			value := CardValue(analyzer.trios[i][0])
			if value > cardValue {
				melds = append(melds, analyzer.trios[i])
			}
		}
	}
	melds = ReSortHint(Trio, melds, cards)
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于三带一的牌
func GetGreaterThanTrioSolo(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	if analyzer.hasTrio {
		triosLen := len(analyzer.trios)
		for i := triosLen - 1; i > -1; i-- {
			cardValue2 := CardValue(analyzer.trios[i][0])
			if cardValue2 > cardValue {
				valueMap := GetCardValueMap(analyzer.trios[i])
				remain := RemoveCardByValue(cards, valueMap)
				solo := GetSoloByCount(remain)
				if len(solo) == 1 {
					trioSolo := append([]int{}, analyzer.trios[i]...)
					trioSolo = append(trioSolo, solo...)
					melds = append(melds, trioSolo)
				}
			}
		}
	}
	melds = ReSortHint(TrioSolo, melds, cards)
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于三带一对的牌
func GetGreaterThanTrioPair(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	if analyzer.hasTrio {
		triosLen := len(analyzer.trios)
		for i := triosLen - 1; i > -1; i-- {
			value := CardValue(analyzer.trios[i][0])
			if value > cardValue {
				valueMap := GetCardValueMap(analyzer.trios[i])
				remain := RemoveCardByValue(cards, valueMap)
				pair := GetPairByCount(remain)
				if len(pair) == 2 {
					trioPair := append([]int{}, analyzer.trios[i]...)
					trioPair = append(trioPair, pair...)
					melds = append(melds, trioPair)
				}
			}
		}
	}
	melds = ReSortHint(TrioPair, melds, cards)
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于顺子的牌
func GetGreaterThanSoloChain(preSoloChain []int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	preSoloChainLen := len(preSoloChain)
	cardValue := CardValue(preSoloChain[0])
	if analyzer.hasSoloChain {
		m, temp := make(map[string]bool), [][]int{}
		for _, soloChain := range analyzer.soloChains {
			key := MeldKey(soloChain)
			if m[key] {
				continue
			}
			m[key] = true
			soloChainLen := len(soloChain)
			for i, card := range soloChain {
				cardValue2 := CardValue(card)
				if i+preSoloChainLen <= soloChainLen && cardValue2 > cardValue {
					temp = append(temp, soloChain[i:i+preSoloChainLen])
				} else {
					break
				}
			}
		}
		tempLen := len(temp)
		for i := tempLen - 1; i > -1; i-- {
			melds = append(melds, temp[i])
		}
	}
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于连对的牌
func GetGreaterThanPairSisters(prePairSisters []int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	prePairSistersLen := len(prePairSisters)
	cardValue := CardValue(prePairSisters[0])
	if analyzer.hasPairSisters {
		m, temp := make(map[string]bool), [][]int{}
		for _, pairSisters := range analyzer.pairSisterss {
			key := MeldKey(pairSisters)
			if m[key] {
				continue
			}
			m[key] = true
			pairSistersLen := len(pairSisters)
			for i := 0; i < pairSistersLen/2; i++ {
				cardValue2 := CardValue(pairSisters[i*2])
				if i*2+prePairSistersLen <= pairSistersLen && cardValue2 > cardValue {
					temp = append(temp, pairSisters[i*2:i*2+prePairSistersLen])
				} else {
					break
				}
			}
		}
		tempLen := len(temp)
		for i := tempLen - 1; i > -1; i-- {
			melds = append(melds, temp[i])
		}
	}
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于飞机的牌
func GetGreaterThanAirplaneChain(preAirplaneChain []int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	preAirplaneChainLen := len(preAirplaneChain)
	cardValue := CardValue(preAirplaneChain[0])
	if analyzer.hasAirplaneChain {
		temp := [][]int{}
		for _, airplaneChain := range analyzer.airplaneChains {
			airplaneChainLen := len(airplaneChain)
			for i := 0; i < airplaneChainLen/3; i++ {
				cardValue2 := CardValue(airplaneChain[i*3])
				if i*3+preAirplaneChainLen <= airplaneChainLen && cardValue2 > cardValue {
					temp = append(temp, airplaneChain[i*3:i*3+preAirplaneChainLen])
				} else {
					break
				}
			}
		}
		tempLen := len(temp)
		for i := tempLen - 1; i > -1; i-- {
			melds = append(melds, temp[i])
		}
	}
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于飞机带小翼的牌
func GetGreaterThanTrioSoloAirplane(preTrioSoloAirplane []int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	preTrioSoloAirplaneLen := len(preTrioSoloAirplane)
	preAirplaneChainLen := preTrioSoloAirplaneLen / 4 * 3
	cardValue := CardValue(preTrioSoloAirplane[0])
	if analyzer.hasAirplaneChain {
		temp := [][]int{}
		for _, airplaneChain := range analyzer.airplaneChains {
			airplaneChainLen := len(airplaneChain)
			for i := 0; i < airplaneChainLen/3; i++ {
				cardValue2 := CardValue(airplaneChain[i*3])
				if i*3+preAirplaneChainLen <= airplaneChainLen && cardValue2 > cardValue {
					valueMap := GetCardValueMap(airplaneChain[i*3 : i*3+preAirplaneChainLen])
					remain := RemoveCardByValue(cards, valueMap)
					remainLen := len(remain)
					if remainLen >= preTrioSoloAirplaneLen/4 {
						trioSoloAirplane := append([]int{}, airplaneChain[i*3:i*3+preAirplaneChainLen]...)
						trioSoloAirplane = append(trioSoloAirplane, remain[remainLen-preTrioSoloAirplaneLen/4:]...)
						temp = append(temp, trioSoloAirplane)
					}
				} else {
					break
				}
			}
		}
		tempLen := len(temp)
		for i := tempLen - 1; i > -1; i-- {
			melds = append(melds, temp[i])
		}
	}
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于飞机带大翼的牌
func GetGreaterThanTrioPairChain(preTrioPairChain []int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	preTrioPairChainLen := len(preTrioPairChain)
	preAirplaneChainLen := preTrioPairChainLen / 5 * 3
	cardValue := CardValue(preTrioPairChain[0])
	temp, tempLen := [][]int{}, 0
	if !analyzer.hasAirplaneChain {
		goto END
	}
	for _, airplaneChain := range analyzer.airplaneChains {
		airplaneChainLen := len(airplaneChain)
		for i := 0; i < airplaneChainLen/3; i++ {
			cardValue2 := CardValue(airplaneChain[i*3])
			if i*3+preAirplaneChainLen <= airplaneChainLen && cardValue2 > cardValue {
				valueMap := GetCardValueMap(airplaneChain[i*3 : i*3+preAirplaneChainLen])
				remain := RemoveCardByValue(cards, valueMap)

				analyzer2 := new(LandlordAnalyzer)
				analyzer2.Analyze(remain)
				if !analyzer2.hasPair {
					continue
				}
				m, pairs := make(map[int]bool), [][]int{}
				pairsLen := len(analyzer2.pairs)
				for i := pairsLen - 1; i > -1; i-- {
					pair := analyzer2.pairs[i]
					if !m[CardValue(pair[0])] {
						pairs = append(pairs, pair)
						m[CardValue(pair[0])] = true
					}
				}
				pairs = ReSortHint(Pair, pairs, remain)
				pairsLen = len(pairs)
				if pairsLen >= preTrioPairChainLen/5 {
					trioPairChain := append([]int{}, airplaneChain[i*3:i*3+preAirplaneChainLen]...)
					pairChain := []int{}
					for _, pair := range pairs[:preTrioPairChainLen/5] {
						pairChain = append(pairChain, pair...)
					}
					trioPairChain = append(trioPairChain, pairChain...)
					temp = append(temp, trioPairChain)
				}
			} else {
				break
			}
		}
	}
	tempLen = len(temp)
	for i := tempLen - 1; i > -1; i-- {
		melds = append(melds, temp[i])
	}
END:
	bombs := GetBombs(cards)
	melds = append(melds, bombs...)
	return melds
}

// 获取大于炸弹的牌
func GetGreaterThanBomb(cardValue int, cards []int) [][]int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	melds := [][]int{}
	if analyzer.hasBomb {
		bombsLen := len(analyzer.bombs)
		for i := bombsLen - 1; i > -1; i-- {
			cardValue2 := CardValue(analyzer.bombs[i][0])
			if cardValue2 > cardValue {
				melds = append(melds, analyzer.bombs[i])
			}
		}
	}
	if analyzer.hasKingBomb {
		melds = append(melds, []int{53, 52})
	}
	return melds
}

// 获取单张
func GetSoloByCount(cards []int) []int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	remain := append([]int{}, cards...)
	if analyzer.hasSoloChain {
		for _, soloChain := range analyzer.soloChains {
			remain = utils.Remove(remain, soloChain)
		}
	}
	cardValueMarker := make(map[int]int)
	for _, card := range remain {
		cardValueMarker[CardValue(card)]++
	}
	count := 1
NEXT:
	for cardValue := 3; cardValue < 18; cardValue++ {
		if cardValueMarker[cardValue] == count {
			temp := GetCardsByValue(remain, cardValue)
			return temp[:1]
		}
	}
	count++
	if count < 5 {
		goto NEXT
	}
	if analyzer.hasSoloChain {
		temp := utils.Remove(cards, remain)
		sort.Ints(temp)
		return temp[:1]
	}
	return []int{}
}

// 获取对子
func GetPairByCount(cards []int) []int {
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)

	count := 2
NEXT:
	for cardValue := 3; cardValue < 16; cardValue++ {
		if analyzer.cardValueMarker[cardValue] == count {
			temp := GetCardsByValue(analyzer.cards, cardValue)
			return temp[:2]
		}
	}
	count++
	if count < 5 {
		goto NEXT
	}
	return []int{}
}
