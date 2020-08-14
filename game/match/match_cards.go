package match

import (
	"ddz/game/ddz"
	"ddz/game/poker"
	"strconv"

	"github.com/szxby/tools/log"
)

// 转换成体总的手牌编码
func changeCardsToSportsCenter(cards []int) string {
	log.Debug("change cards:%v", cards)
	var ret string
	for n, c := range cards {
		value := poker.CardValue(c)
		cardType := poker.CardType[c]
		// 大小王
		if cardType == poker.JokerCard {
			ret += strconv.Itoa(c)
		} else {
			ret += strconv.Itoa((cardType-1)*13 + value - 3)
		}
		if n != len(cards)-1 {
			ret += ","
		}
	}
	log.Debug("finish change:%v", ret)
	return ret
}

// 获取春天的string
func getSpring(spring, lSpring int) string {
	if spring == 1 || lSpring == 1 {
		return "1"
	}
	return "0"
}

// 获取加倍的string
func getDouble(double uint) string {
	if double > 1 {
		return "1"
	}
	return "0"
}

// 获取同桌选手id
func getTablePlayerID(data map[int]*ddz.LandlordMatchPlayerData, uid int) string {
	var ret string
	for id, p := range data {
		if id == uid {
			continue
		}
		ret += strconv.Itoa(p.User.BaseData.UserData.AccountID) + ","
	}
	ret = ret[:len(ret)-1]
	return ret
}
