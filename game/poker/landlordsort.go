package poker

import (
	"ddz/utils"
)

func ReSortLandlordCards(cards []int) []int {
	cardsType := GetLandlordCardsType(cards)
	switch cardsType {
	// 三带一、三带一对、飞机、飞机带小翼、飞机带大翼、四带两单、四带两对
	case TrioSolo, TrioPair, AirplaneChain, TrioSoloAirplane, TrioPairChain, FourDualsolo, FourDualpair:
		newCards := []int{}
		for _, card := range cards {
			if CountCardValue(cards, card) > 2 {
				newCards = append(newCards, card)
			}
		}
		newCards = append(newCards, utils.Remove(cards, newCards)...)
		return newCards
	}
	return cards
}

func ReSortHint(hintType int, melds [][]int, cards []int) [][]int {
	if len(melds) == 0 {
		return [][]int{}
	}
	cardValueMarker := make(map[int]int)
	for _, card := range cards {
		cardValueMarker[CardValue(card)]++
	}
	newMelds := [][]int{}
	m := make(map[int]bool)
	var count int
	switch hintType {
	case Solo:
		count = 1
	case Pair:
		count = 2
	case Trio, TrioSolo, TrioPair:
		count = 3
	}
NEXT:
	for _, meld := range melds {
		value := CardValue(meld[0])
		if !m[value] && cardValueMarker[value] == count {
			newMelds = append(newMelds, meld)
			m[value] = true
		}
	}
	count++
	if count < 5 {
		goto NEXT
	}
	return newMelds
}
