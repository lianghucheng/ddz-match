package poker

import (
	"ddz/utils"
	"sort"
)

type LandlordAnalyzer struct {
	cardValueMarker map[int]int // 记牌器(Key 牌值、Value 牌的数量)
	cards           []int

	hasKingBomb      bool
	hasBomb          bool
	hasTrio          bool
	hasPair          bool
	hasAirplaneChain bool
	hasPairSisters   bool
	hasSoloChain     bool

	kingBomb       []int
	bombs          [][]int
	trios          [][]int
	pairs          [][]int
	airplaneChains [][]int
	pairSisterss   [][]int
	soloChains     [][]int

	unrelated []int
}

func (analyzer *LandlordAnalyzer) init() {
	analyzer.cardValueMarker = make(map[int]int)
	analyzer.cards = []int{}

	analyzer.hasKingBomb = false
	analyzer.hasBomb = false
	analyzer.hasTrio = false
	analyzer.hasPair = false
	analyzer.hasAirplaneChain = false
	analyzer.hasPairSisters = false
	analyzer.hasSoloChain = false

	analyzer.kingBomb = []int{}
	analyzer.bombs = [][]int{}
	analyzer.trios = [][]int{}
	analyzer.pairs = [][]int{}
	analyzer.airplaneChains = [][]int{}
	analyzer.pairSisterss = [][]int{}
	analyzer.soloChains = [][]int{}

	analyzer.unrelated = []int{}
}

func (analyzer *LandlordAnalyzer) setCardValueMarker() {
	for i := 3; i < 18; i++ { // 3、4、5、6、7、8、9、10、J、Q、K、A、2、小王、大王
		analyzer.cardValueMarker[i] = 0
	}
	for _, card := range analyzer.cards {
		analyzer.cardValueMarker[CardValue(card)]++
	}
}

// 传入的牌必须降序排列
func (analyzer *LandlordAnalyzer) Analyze(cards []int) {
	analyzer.init()
	analyzer.cards = append([]int{}, cards...)
	analyzer.setCardValueMarker()
	analyzer.analyzeKingBomb()
	analyzer.analyzeBomb()
	analyzer.analyzeTrio()
	if analyzer.hasTrio {
		analyzer.analyzeAirplaneChains(unite(analyzer.trios))
	}
	analyzer.analyzePair()
	if analyzer.hasPair {
		analyzer.analyzePairSisters(unite(analyzer.pairs))
	}
	analyzer.analyzeSoloChain(exclude(analyzer.cards))
	analyzer.analyzeUnrelated()
}

func (analyzer *LandlordAnalyzer) analyzeKingBomb() {
	if len(analyzer.cards) > 1 && analyzer.cards[0] == 53 && analyzer.cards[1] == 52 {
		analyzer.hasKingBomb = true
		analyzer.kingBomb = []int{53, 52}
	} else {
		analyzer.hasKingBomb = false
	}
}

func (analyzer *LandlordAnalyzer) analyzeBomb() {
	for cardValue := 15; cardValue > 2; cardValue-- {
		if analyzer.cardValueMarker[cardValue] == 4 {
			bomb := []int{4*cardValue - 12, 4*cardValue - 11, 4*cardValue - 10, 4*cardValue - 9}
			analyzer.bombs = append(analyzer.bombs, bomb)
		}
	}
	if len(analyzer.bombs) == 0 {
		analyzer.hasBomb = false
	} else {
		analyzer.hasBomb = true
	}
}

func (analyzer *LandlordAnalyzer) analyzeTrio() {
	for value := 15; value > 2; value-- {
		if analyzer.cardValueMarker[value] > 2 {
			temp := GetCardsByValue(analyzer.cards, value)
			analyzer.trios = append(analyzer.trios, temp[:3])
		}
	}
	if len(analyzer.trios) == 0 {
		analyzer.hasTrio = false
	} else {
		analyzer.hasTrio = true
	}
}

func (analyzer *LandlordAnalyzer) analyzePair() {
	for value := 15; value > 2; value-- {
		countCardValue := analyzer.cardValueMarker[value]
		switch countCardValue {
		case 2, 3:
			temp := GetCardsByValue(analyzer.cards, value)
			analyzer.pairs = append(analyzer.pairs, temp[:2])
		case 4:
			temp := GetCardsByValue(analyzer.cards, value)
			analyzer.pairs = append(analyzer.pairs, temp[:2])
			analyzer.pairs = append(analyzer.pairs, temp[2:])
		}
	}
	if len(analyzer.pairs) == 0 {
		analyzer.hasPair = false
	} else {
		analyzer.hasPair = true
	}
}

func (analyzer *LandlordAnalyzer) analyzeAirplaneChains(cards []int) {
	airplaneChains := GetLandlordAirplaneChains(cards)
	if len(airplaneChains) == 0 {
		analyzer.hasAirplaneChain = false
	} else {
		analyzer.hasAirplaneChain = true
		for _, airplaneChain := range airplaneChains {
			analyzer.airplaneChains = append(analyzer.airplaneChains, airplaneChain)
		}
	}
}

func (analyzer *LandlordAnalyzer) analyzePairSisters(cards []int) {
	remain := append([]int{}, cards...)
	pairSisterss := GetLandlordPairSisters(remain)
	if len(pairSisterss) == 0 {
		if len(analyzer.pairSisterss) == 0 {
			analyzer.hasPairSisters = false
		}
	} else {
		analyzer.hasPairSisters = true
		for _, pairSisters := range pairSisterss {
			analyzer.pairSisterss = append(analyzer.pairSisterss, pairSisters)
			remain = utils.Remove(remain, pairSisters)
		}
		analyzer.analyzePairSisters(remain)
	}
}

func (analyzer *LandlordAnalyzer) analyzeSoloChain(cards []int) {
	remain := append([]int{}, cards...)
	soloChains := GetLandlordSoloChains(remain)
	if len(soloChains) == 0 {
		if len(analyzer.soloChains) == 0 {
			analyzer.hasSoloChain = false
		}
	} else {
		analyzer.hasSoloChain = true
		for _, soloChain := range soloChains {
			analyzer.soloChains = append(analyzer.soloChains, soloChain)
			remain = utils.Remove(remain, soloChain)
		}
		analyzer.analyzeSoloChain(remain)
	}
}

func (analyzer *LandlordAnalyzer) analyzeUnrelated() {
	cardMarker := make(map[int]bool)
	cardMarker = GetCardsMap(analyzer.kingBomb, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.bombs, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.trios, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.pairs, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.airplaneChains, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.pairSisterss, cardMarker)
	cardMarker = GetMeldsCardsMap(analyzer.soloChains, cardMarker)
	for _, card := range analyzer.cards {
		if !cardMarker[card] {
			analyzer.unrelated = append(analyzer.unrelated, card)
			cardMarker[card] = true
		}
	}
	sort.Ints(analyzer.unrelated)
}

func (analyzer *LandlordAnalyzer) Print() {
	melds := [][]int{}
	if analyzer.hasKingBomb {
		melds = append(melds, analyzer.kingBomb)
	}
	melds = append(melds, analyzer.bombs...)
	melds = append(melds, analyzer.airplaneChains...)
	melds = append(melds, analyzer.trios...)
	melds = append(melds, analyzer.pairSisterss...)
	melds = append(melds, analyzer.pairs...)
	melds = append(melds, analyzer.soloChains...)
	if len(analyzer.unrelated) > 0 {
		melds = append(melds, analyzer.unrelated)
	}
}

func unite(melds [][]int) []int {
	temp := []int{}
	for _, meld := range melds {
		if CardValue(meld[0]) < 15 {
			temp = append(temp, meld...)
		}
	}
	return temp
}

// 把大于A的牌排除
func exclude(cards []int) []int {
	temp := []int{}
	for _, card := range cards {
		if CardValue(card) < 15 {
			temp = append(temp, card)
		}
	}
	return temp
}

// 排除顺子、对子
func (analyzer *LandlordAnalyzer) analyze2(cards []int) {
	analyzer.init()
	analyzer.cards = append([]int{}, cards...)
	analyzer.setCardValueMarker()
	//analyzer.analyzeKingBomb()
	//if analyzer.hasKingBomb {
	//	analyzer.cards = utils.Remove(analyzer.cards, []int{53, 52})
	//	analyzer.setCardValueMarker()
	//}
	//analyzer.analyzeBomb()
	//if analyzer.hasBomb {
	//	for _, bomb := range analyzer.bombs {
	//		analyzer.cards = utils.Remove(analyzer.cards, bomb)
	//	}
	//	analyzer.setCardValueMarker()
	//}
	analyzer.analyzeSoloChain(exclude(analyzer.cards))
	if analyzer.hasSoloChain {
		for _, soloChain := range analyzer.soloChains {
			analyzer.cards = utils.Remove(analyzer.cards, soloChain)
		}
		analyzer.setCardValueMarker()
	}
	analyzer.analyzeTrio()
	if analyzer.hasTrio {
		for _, trio := range analyzer.trios {
			analyzer.cards = utils.Remove(analyzer.cards, trio)
		}
		analyzer.setCardValueMarker()
		// analyzer.analyzeAirplaneChains(unite(analyzer.trios))
	}
	analyzer.analyzePair()
	if analyzer.hasPair {
		for _, pair := range analyzer.pairs {
			analyzer.cards = utils.Remove(analyzer.cards, pair)
		}
		analyzer.setCardValueMarker()
		// analyzer.analyzePairSisters(unite(analyzer.pairs))
	}
	analyzer.unrelated = append([]int{}, analyzer.cards...)
	sort.Ints(analyzer.unrelated)
}

func (analyzer *LandlordAnalyzer) GetMinDiscards(cards []int) []int {
	/*
			analyzer.init()
			analyzer.cards = append([]int{}, cards...)
			analyzer.setCardValueMarker()

					analyzer.analyzeKingBomb()
					if analyzer.hasKingBomb {
						analyzer.cards = utils.Remove(analyzer.cards, analyzer.kingBomb)
						analyzer.setCardValueMarker()
						if len(analyzer.cards) == 0 {
							return []int{analyzer.kingBomb[1]}
						}
					}
					analyzer.analyzeBomb()
					if analyzer.hasBomb {
						for _, bomb := range analyzer.bombs {
							analyzer.cards = utils.Remove(analyzer.cards, bomb)
						}
						analyzer.setCardValueMarker()
						if len(analyzer.cards) == 0 {
							return analyzer.bombs[len(analyzer.bombs)-1]
						}
					}
					analyzer.analyzeTrio()
					if analyzer.hasTrio {
						analyzer.analyzeAirplaneChains(unite(analyzer.trios))
					}
					analyzer.analyzePair()
					if analyzer.hasPair {
						analyzer.analyzePairSisters(unite(analyzer.pairs))
					}
				analyzer.analyzeSoloChain(exclude(analyzer.cards))
				analyzer.analyzeUnrelated()
				analyzer.Print()

				minCard := analyzer.cards[len(analyzer.cards)-1]
					if analyzer.hasAirplaneChain {
						meld := analyzer.airplaneChains[len(analyzer.airplaneChains)-1]
						if utils.InArray(meld, minCard) {
							remain := RemoveCardByValue(analyzer.cards, GetCardValueMap(meld))
							return mixAirplane(meld, exclude(remain))
						}
					}
					if analyzer.hasTrio {
						meld := analyzer.trios[len(analyzer.trios)-1]
						if utils.InArray(meld, minCard) {
							remain := RemoveCardByValue(analyzer.cards, GetCardValueMap(meld))
							return mixTrio(meld, exclude(remain))
						}
					}
					if analyzer.hasPairSisters {
						meld := analyzer.pairSisterss[len(analyzer.pairSisterss)-1]
						if utils.InArray(meld, minCard) {
							return meld
						}
					}
					if analyzer.hasSoloChain {
						meld := analyzer.soloChains[len(analyzer.soloChains)-1]
						if utils.InArray(meld, minCard) {
							return meld
						}
					}
					if analyzer.hasPair {
						meld := analyzer.pairs[len(analyzer.pairs)-1]
						if utils.InArray(meld, minCard) {
							remain := RemoveCardByValue(analyzer.cards, GetCardValueMap(meld))
							return mixPair(meld, exclude(remain))
						}
					}
		meld := []int{minCard}
		remain := RemoveCardByValue(analyzer.cards, GetCardValueMap(meld))
		return mixSolo(meld, exclude(remain))
	*/
	return []int{cards[len(cards)-1]}
}

func mixAirplane(meld []int, cards []int) []int {
	kickerLen := len(meld) / 3
	if kickerLen > len(cards) {
		return meld
	}
	analyzer := new(LandlordAnalyzer)
	analyzer.analyze2(cards)
	if len(analyzer.unrelated) >= kickerLen {
		newMeld := append([]int{}, meld...)
		newMeld = append(newMeld, analyzer.unrelated[:kickerLen]...)
		return newMeld
	}
	if len(analyzer.pairs) >= kickerLen {
		newMeld := append([]int{}, meld...)
		newMeld = append(newMeld, unite(analyzer.pairs[len(analyzer.pairs)-kickerLen:])...)
		return newMeld
	}
	return meld
}

func mixTrio(meld []int, cards []int) []int {
	if len(cards) == 0 {
		return meld
	}
	analyzer := new(LandlordAnalyzer)
	analyzer.analyze2(cards)
	if len(analyzer.unrelated) > 0 {
		newMeld := append([]int{}, meld...)
		newMeld = append(newMeld, analyzer.unrelated[0])
		return newMeld
	}
	if len(analyzer.pairs) > 0 {
		newMeld := append([]int{}, meld...)
		newMeld = append(newMeld, analyzer.pairs[len(analyzer.pairs)-1]...)
		return newMeld
	}
	return meld
}

func mixPair(meld []int, cards []int) []int {
	if len(cards) == 0 {
		return meld
	}
	analyzer := new(LandlordAnalyzer)
	analyzer.Analyze(cards)
	if len(analyzer.airplaneChains) > 0 {
		airplaneChain := analyzer.airplaneChains[len(analyzer.airplaneChains)-1]
		newCards := append([]int{}, cards...)
		newCards = append(newCards, meld...)
		sort.Sort(sort.Reverse(sort.IntSlice(newCards)))
		remain := utils.Remove(newCards, airplaneChain)
		return mixAirplane(airplaneChain, remain)
	}
	if len(analyzer.trios) > 0 {
		newMeld := append([]int{}, analyzer.trios[len(analyzer.trios)-1]...)
		newMeld = append(newMeld, meld...)
		return newMeld
	}
	return meld
}

func mixSolo(meld []int, cards []int) []int {
	return mixPair(meld, cards)
}
