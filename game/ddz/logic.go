package ddz

import (
	"ddz/conf"
	. "ddz/game/player"
	"ddz/game/poker"
	"ddz/msg"
	"ddz/utils"
	"fmt"
	"sort"
	"time"

	"github.com/name5566/leaf/log"
)

func (game *LandlordMatchRoom) score(userID int) {
	//最高三分
	playerData := game.userIDPlayerDatas[userID]
	playerData.state = landlordActionBid
	score := make([]int, 0)
	switch game.maxscore {
	case 0:
		{
			score = append(score, []int{0, 1, 2, 3}...)
		}
	case 1:
		{
			score = append(score, []int{0, 2, 3}...)
		}
	case 2:
		{
			score = append(score, []int{0, 3}...)
		}
	}
	game.broadcast(&msg.S2C_ActionLandlordBid{
		Position:  playerData.position,
		Countdown: conf.GetCfgTimeout().LandlordBid,
		Score:     score,
	}, game.PositionUserIDs, -1)

	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 叫分", userID)
	game.bidTimer = skeleton.AfterFunc((time.Duration(conf.GetCfgTimeout().LandlordBid+2))*time.Second, func() {
		log.Debug("userID %v 自动叫0", userID)
		game.doscore(userID, 0)
	})
}

func (game *LandlordMatchRoom) doscore(userID int, score int) {
	playerData := game.userIDPlayerDatas[userID]
	if playerData.state != landlordActionBid {
		return
	}
	if score != 0 && score <= game.maxscore {
		return
	}
	game.bidTimer.Stop()
	playerData.state = landlordWaiting

	game.broadcast(&msg.S2C_LandlordBid{
		Position: playerData.position,
		Score:    score,
	}, game.PositionUserIDs, -1)
	log.Debug("玩家%v叫分%v", userID, score)
	dealerPlayerData := game.userIDPlayerDatas[game.dealerUserID]
	nextUserID := game.PositionUserIDs[(playerData.position+1)%game.rule.MaxPlayers]
	lastPos := (dealerPlayerData.position + game.rule.MaxPlayers - 1) % game.rule.MaxPlayers
	playerData.score = score
	if score > game.maxscore {
		game.maxscore = score
	}

	if score == 3 {
		skeleton.AfterFunc(1*time.Second, func() {
			game.decideLandlord(userID)
		})
		return
	}
	if playerData.position == lastPos {
		//比较叫分的大小决定谁是地主
		max := game.userIDPlayerDatas[game.dealerUserID].score
		userID := game.dealerUserID
		for i := 1; i < len(game.userIDPlayerDatas); i++ {
			position := ((game.userIDPlayerDatas[userID].position) + 1) % game.rule.MaxPlayers
			nextUserID := game.PositionUserIDs[position]
			if game.userIDPlayerDatas[nextUserID].score <= max {
				continue
			}
			max = game.userIDPlayerDatas[nextUserID].score
			userID = nextUserID

		}
		skeleton.AfterFunc(1*time.Second, func() {
			game.decideLandlord(userID)
		})
	} else {
		game.score(nextUserID)
	}
}

// 确定地主
func (game *LandlordMatchRoom) decideLandlord(userID int) {
	game.broadcast(&msg.S2C_ClearAction{}, game.PositionUserIDs, -1)

	game.landlordUserID = userID
	playerData := game.userIDPlayerDatas[game.landlordUserID]
	for i := 1; i < game.rule.MaxPlayers; i++ {
		peasantUserID := game.PositionUserIDs[(playerData.position+i)%game.rule.MaxPlayers]
		game.peasantUserIDs = append(game.peasantUserIDs, peasantUserID)
	}
	//确定庄家以后，更新玩家的公共分
	for i := 0; i < len(game.PositionUserIDs); i++ {
		score := 1
		if game.userIDPlayerDatas[game.landlordUserID].score == 0 {
			score *= game.rule.BaseScore
		} else {
			score = game.userIDPlayerDatas[game.landlordUserID].score * game.rule.BaseScore
		}
		game.userIDPlayerDatas[game.PositionUserIDs[i]].DealerScore = score
		game.userIDPlayerDatas[game.PositionUserIDs[i]].Public = score

		game.sendRoomPanel(game.PositionUserIDs[i])
		game.gameRecords[userID].Result[game.count-1].Bottom = score

	}

	game.broadcast(&msg.S2C_DecideLandlord{
		Position: playerData.position,
	}, game.PositionUserIDs, -1)
	// 最后三张
	game.lastThree = game.rests[:3]
	game.rests = []int{}
	sort.Sort(sort.Reverse(sort.IntSlice(game.lastThree)))
	log.Debug("三张: %v", poker.ToCardsString(game.lastThree))

	game.broadcast(&msg.S2C_UpdateLandlordLastThree{
		Cards: game.lastThree,
	}, game.PositionUserIDs, -1)

	playerData.hands = append(playerData.hands, game.lastThree...)
	sort.Sort(sort.Reverse(sort.IntSlice(playerData.hands)))

	if playerData, ok := game.userIDPlayerDatas[userID]; ok {
		playerData.user.WriteMsg(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
	}
	game.broadcast(&msg.S2C_UpdatePokerHands{
		Position:      playerData.position,
		Hands:         []int{},
		NumberOfHands: len(playerData.hands),
	}, game.PositionUserIDs, playerData.position)

	skeleton.AfterFunc(1*time.Second, func() {
		game.double()
	})
}

// 加倍
func (game *LandlordMatchRoom) double() {
	actionTimestamp := time.Now().Unix()
	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
		playerData.state = landlordActionDouble
		playerData.actionTimestamp = actionTimestamp

		if playerData, ok := game.userIDPlayerDatas[userID]; ok {
			playerData.user.WriteMsg(&msg.S2C_ActionLandlordDouble{
				Countdown: conf.GetCfgTimeout().LandlordDouble,
			})
		}
	}
	log.Debug("等待所有人加倍")
	game.doubleTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordDouble+2)*time.Second, func() {
		for _, userID := range game.PositionUserIDs {
			playerData := game.userIDPlayerDatas[userID]
			if playerData.state == landlordActionDouble {
				log.Debug("userID %v 自动不加倍", userID)
				game.doDouble(userID, false)
			}
		}
	})
}

func (game *LandlordMatchRoom) doDouble(userID int, double bool) {
	playerData := game.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDouble {
		return
	}
	playerData.state = landlordWaiting
	lable := 1
	if double {
		lable = 2
	}
	game.gameRecords[userID].Result[game.count-1].Multiple = lable
	game.gameRecords[userID].Result[game.count-1].ThreeCards = game.lastThree
	game.broadcast(&msg.S2C_LandlordDouble{
		Position: playerData.position,
		Double:   double,
	}, game.PositionUserIDs, -1)
	if userID == game.landlordUserID {
		for i := 0; i < len(game.PositionUserIDs); i++ {
			if game.PositionUserIDs[i] != game.landlordUserID {
				if double {
					game.userIDPlayerDatas[game.PositionUserIDs[i]].Dealer = 2
				} else {
					game.userIDPlayerDatas[game.PositionUserIDs[i]].Dealer = 1
				}
				game.sendRoomPanel(game.PositionUserIDs[i])
			}
		}
		if double {
			playerData.Dealer = 2
		} else {
			playerData.Dealer = 1
		}
		game.sendRoomPanel(userID)
	}
	if userID != game.landlordUserID {
		for i := 0; i < len(game.PositionUserIDs); i++ {
			if game.PositionUserIDs[i] == game.landlordUserID {
				if double {
					game.userIDPlayerDatas[game.PositionUserIDs[i]].Xian += 2
				} else {
					game.userIDPlayerDatas[game.PositionUserIDs[i]].Xian += 1
				}
				game.sendRoomPanel(game.PositionUserIDs[i])
			}
		}
		if double {
			playerData.Xian = 2
		} else {
			playerData.Xian = 1
		}
		game.sendRoomPanel(userID)
	}

	if game.allWaiting() {
		game.doubleTimer.Stop()
		skeleton.AfterFunc(1500*time.Millisecond, func() {
			game.broadcast(&msg.S2C_ClearAction{}, game.PositionUserIDs, -1)
			game.discard(game.landlordUserID, poker.ActionLandlordDiscardMust)
		})
	}
}

// 出牌
func (game *LandlordMatchRoom) discard(userID int, actionDiscardType int) {
	playerData := game.userIDPlayerDatas[userID]
	playerData.state = landlordActionDiscard
	playerData.actionDiscardType = actionDiscardType

	game.broadcast(&msg.S2C_ActionLandlordDiscard{
		Position:  playerData.position,
		Countdown: conf.GetCfgTimeout().LandlordDiscard,
	}, game.PositionUserIDs, playerData.position)
	playerData.discardTimeStamp = time.Now().UnixNano() / 1e6
	prevDiscards := []int{}
	countdown := conf.GetCfgTimeout().LandlordDiscard
	hint := make([][]int, 0)
	switch playerData.actionDiscardType {
	case poker.ActionLandlordDiscardNothing:
		if playerData.hosted {
			goto HOST
		}
		countdown = conf.GetCfgTimeout().LandlordDiscardNothing
	case poker.ActionLandlordDiscardAlternative:
		if playerData.hosted {
			goto HOST
		}
		discarderPlayerData := game.userIDPlayerDatas[game.discarderUserID]
		prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		if !poker.CompareLandlordHands(prevDiscards, playerData.hands) {
			hint = poker.GetDiscardHint(prevDiscards, playerData.hands)
			log.Debug("提示出牌:%v", poker.ToMeldsString(hint))
		}

	case poker.ActionLandlordDiscardMust:
		if playerData.hosted {
			goto HOST
		}
	}
	if playerData, ok := game.userIDPlayerDatas[userID]; ok {
		playerData.user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
			ActionDiscardType: playerData.actionDiscardType,
			Position:          playerData.position,
			Countdown:         countdown,
			PrevDiscards:      prevDiscards,
			Hint:              hint,
		})
	}
	playerData.actionTimestamp = time.Now().Unix()

	log.Debug("等待 userID %v 出牌 动作: %v", userID, playerData.actionDiscardType)
	game.discardTimer = skeleton.AfterFunc(time.Duration(countdown+2)*time.Second, func() {
		switch playerData.actionDiscardType {
		case poker.ActionLandlordDiscardNothing:
			log.Debug("userID %v 自动不出", userID)
			game.doDiscard(userID, []int{})
		default:
			playerData.count++
			if playerData.count >= 2 {
				playerData.hosted = true
				playerData.user.WriteMsg(&msg.S2C_ClearAction{})
				playerData.user.WriteMsg(&msg.S2C_SystemHost{
					Position: playerData.position,
					Host:     true,
				})
			}
			game.doHostDiscard(userID)
		}
	})
	return
HOST: // 托管出牌
	skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordSystemHost)*time.Second, func() {
		game.doHostDiscard(userID)
	})
	return
}
func (game *LandlordMatchRoom) doDiscard(userID int, cards []int) {
	playerData := game.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDiscard {
		return
	}
	cards = poker.ReSortLandlordCards(cards)
	cardsLen := len(cards)
	cardsType := poker.GetLandlordCardsType(cards)
	contain := utils.Contain(playerData.hands, cards)

	var prevDiscards []int
	if game.discarderUserID > 0 && game.discarderUserID != userID {
		discarderPlayerData := game.userIDPlayerDatas[game.discarderUserID]
		prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
	}
	if cardsLen == 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardMust ||
		cardsLen > 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardNothing ||
		cardsLen > 0 && !contain || cardsLen > 0 && cardsType == poker.Error ||
		cardsLen > 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardAlternative && !poker.CompareLandlordDiscard(cards, prevDiscards) {
		if playerData, ok := game.userIDPlayerDatas[userID]; ok {
			after := int(time.Now().Unix() - playerData.actionTimestamp)

			countdown := conf.GetCfgTimeout().LandlordDiscard - after
			if countdown > 1 {
				playerData.user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
					ActionDiscardType: playerData.actionDiscardType,
					Position:          playerData.position,
					Countdown:         countdown - 1,
					PrevDiscards:      prevDiscards,
				})
			}
		}
		return
	}
	if game.discardTimer != nil {
		game.discardTimer.Stop()
		game.discardTimer = nil
	}
	playerData.state = landlordWaiting

	game.broadcast(&msg.S2C_LandlordDiscard{
		Position: playerData.position,
		Cards:    cards,
		CardType: cardsType,
	}, game.PositionUserIDs, -1)
	playerData.costTimeBydiscard += time.Now().UnixNano()/1e6 - playerData.discardTimeStamp
	nextUserID := game.PositionUserIDs[(playerData.position+1)%game.rule.MaxPlayers]
	if cardsLen == 0 {
		log.Debug("userID %v 不出", userID)
		if game.discarderUserID == nextUserID {
			game.discard(nextUserID, poker.ActionLandlordDiscardMust)
		} else {
			nextUserPlayerData := game.userIDPlayerDatas[nextUserID]
			if poker.CompareLandlordHands(prevDiscards, nextUserPlayerData.hands) {
				game.discard(nextUserID, poker.ActionLandlordDiscardNothing)
			} else {
				if nextUserPlayerData.hosted {
					game.discard(nextUserID, poker.ActionLandlordDiscardNothing)
					return
				}
				game.discard(nextUserID, poker.ActionLandlordDiscardAlternative)
			}
		}
		return
	}
	switch cardsType {

	case poker.KingBomb, poker.Bomb:
		for userID, player := range game.userIDPlayerDatas {
			player.Boom++
			player.Public *= 2
			game.sendRoomPanel(userID)
		}
	default:

	}
	game.discarderUserID = userID
	game.discards = append(game.discards, cards...)
	playerData.discards = append(playerData.discards, cards)
	playerData.hands = utils.Remove(playerData.hands, cards)
	log.Debug("userID %v, 出牌: %v", userID, poker.ToCardsString(cards))
	if playerData, ok := game.userIDPlayerDatas[userID]; ok {
		playerData.user.WriteMsg(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
	}
	game.broadcast(&msg.S2C_UpdatePokerHands{
		Position:      playerData.position,
		Hands:         []int{},
		NumberOfHands: len(playerData.hands),
	}, game.PositionUserIDs, playerData.position)

	if len(playerData.hands) == 0 {
		game.winnerUserIDs = append(game.winnerUserIDs, userID)
		skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordEndPrepare)*time.Millisecond, func() {
			game.EndGame()
		})
		return
	}
	if game.discarderUserID == nextUserID {
		game.discard(nextUserID, poker.ActionLandlordDiscardMust)
	} else {
		nextUserPlayerData := game.userIDPlayerDatas[nextUserID]
		if poker.CompareLandlordHands(cards, nextUserPlayerData.hands) {
			game.discard(nextUserID, poker.ActionLandlordDiscardNothing)
		} else {
			if nextUserPlayerData.hosted {
				game.discard(nextUserID, poker.ActionLandlordDiscardNothing)
				return
			}
			game.discard(nextUserID, poker.ActionLandlordDiscardAlternative)
		}
	}
}

// 托管出牌
func (game *LandlordMatchRoom) doHostDiscard(userID int) {
	playerData := game.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDiscard {
		return
	}
	switch playerData.actionDiscardType {
	case poker.ActionLandlordDiscardNothing:
		game.doDiscard(userID, []int{})
		return
	case poker.ActionLandlordDiscardAlternative:
		game.doDiscard(userID, []int{})
		return
	case poker.ActionLandlordDiscardMust:
		analyzer := new(poker.LandlordAnalyzer)
		minCards := analyzer.GetMinDiscards(playerData.hands)
		log.Debug("userID %v 托管出牌: %v", userID, poker.ToCardsString(minCards))
		game.doDiscard(userID, minCards)
		return
	}
}

func (game *LandlordMatchRoom) doSystemHost(userID int, host bool) {
	playerData := game.userIDPlayerDatas[userID]
	if playerData.hosted == host || game.State != roomGame {
		return
	}
	playerData.hosted = host
	playerData.user.WriteMsg(&msg.S2C_SystemHost{
		Position: playerData.position,
		Host:     host,
	})
	if host {
		game.doHostDiscard(userID)
	}
	if !host {
		playerData.count = 0
	}
}

func (game *LandlordMatchRoom) Leave(userID int) {
	game.finisherUserID = -1
	m := msg.C2S_EndMatch{
		Id:      userID,
		MatchId: game.rule.MatchId,
	}
	skeleton.ChanRPCServer.Go("EndMatch", &m)
}

// 断线重连
func (game *LandlordMatchRoom) reconnect(userID int) {
	thePlayerData := game.userIDPlayerDatas[userID]
	if thePlayerData == nil {
		return
	}
	//取消托管
	if thePlayerData.hosted {
		thePlayerData.count = 0
		thePlayerData.user.WriteMsg(&msg.S2C_SystemHost{
			Position: thePlayerData.position,
			Host:     false,
		})
	}
	thePlayerData.user.WriteMsg(&msg.S2C_GameStart{})
	thePlayerData.user.WriteMsg(&msg.S2C_MatchInfo{
		RoundNum:    game.rule.RoundNum,
		Process:     fmt.Sprintf("第%v局 第一幅", game.count),
		Level:       fmt.Sprintf("%v/%v", thePlayerData.Level, game.rule.MaxPlayers),
		Competition: "前三晋级",
	})
	if game.landlordUserID > 0 {
		landlordPlayerData := game.userIDPlayerDatas[game.landlordUserID]
		thePlayerData.user.WriteMsg(&msg.S2C_DecideLandlord{
			Position: landlordPlayerData.position,
		})
		thePlayerData.user.WriteMsg(&msg.S2C_UpdateLandlordLastThree{
			Cards: game.lastThree,
		})
		game.sendRoomPanel(userID)
	}
	if game.discarderUserID > 0 {
		discarderPlayerData := game.userIDPlayerDatas[game.discarderUserID]
		if len(discarderPlayerData.discards) > 1 {
			prevDiscards := discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
			thePlayerData.user.WriteMsg(&msg.S2C_LandlordDiscard{
				Position: discarderPlayerData.position,
				Cards:    prevDiscards,
			})
		}
	}
	game.getPlayerData(thePlayerData.user, thePlayerData, false)

	for i := 1; i < game.rule.MaxPlayers; i++ {
		otherUserID := game.PositionUserIDs[(thePlayerData.position+i)%game.rule.MaxPlayers]
		otherPlayerData := game.userIDPlayerDatas[otherUserID]

		game.getPlayerData(thePlayerData.user, otherPlayerData, true)
	}
}

func (game *LandlordMatchRoom) getPlayerData(user *User, playerData *LandlordMatchPlayerData, other bool) {
	hands := playerData.hands
	if other {
		hands = []int{}
	}
	user.WriteMsg(&msg.S2C_UpdatePokerHands{
		Position:      playerData.position,
		Hands:         hands,
		NumberOfHands: len(playerData.hands),
	})
	if playerData.hosted {
		user.WriteMsg(&msg.S2C_SystemHost{
			Position: playerData.position,
			Host:     true,
		})
	}
	switch playerData.state {
	case landlordActionBid:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordBid - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordBid{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case landlordActionDouble:
		if other {
			return
		}
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordDouble - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordDouble{
				Countdown: countdown - 1,
			})
		}
	case landlordActionDiscard:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		var prevDiscards []int
		if game.discarderUserID > 0 && game.discarderUserID != user.BaseData.UserData.UserID {
			discarderPlayerData := game.userIDPlayerDatas[game.discarderUserID]
			prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		}
		countdown := conf.GetCfgTimeout().LandlordDiscard - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
				ActionDiscardType: playerData.actionDiscardType,
				Position:          playerData.position,
				Countdown:         countdown - 1,
				PrevDiscards:      prevDiscards,
			})
		}
	}
}
func (game *LandlordMatchRoom) players(playerData *LandlordMatchPlayerData, userID int) {
	hands := game.userIDPlayerDatas[userID].hands
	if playerData.user.BaseData.UserData.UserID != userID {
		hands = []int{}
	}
	playerData.user.WriteMsg(&msg.S2C_UpdatePokerHands{
		Position:      game.userIDPlayerDatas[userID].position,
		Hands:         hands,
		NumberOfHands: len(game.userIDPlayerDatas[userID].hands),
	})
	if playerData.hosted {
		playerData.user.WriteMsg(&msg.S2C_SystemHost{
			Position: game.userIDPlayerDatas[userID].position,
			Host:     true,
		})
	}
	switch game.userIDPlayerDatas[userID].state {
	case landlordActionBid:
		after := int(time.Now().Unix() - game.userIDPlayerDatas[userID].actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordBid - after
		if countdown > 1 {
			playerData.user.WriteMsg(&msg.S2C_ActionLandlordBid{
				Position:  game.userIDPlayerDatas[userID].position,
				Countdown: countdown - 1,
			})
		}
	case landlordActionDouble:
		after := int(time.Now().Unix() - game.userIDPlayerDatas[userID].actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordDouble - after
		if countdown > 1 {
			playerData.user.WriteMsg(&msg.S2C_ActionLandlordDouble{
				Countdown: countdown - 1,
			})
		}
	case landlordActionDiscard:
		after := int(time.Now().Unix() - game.userIDPlayerDatas[userID].actionTimestamp)
		var prevDiscards []int
		if game.discarderUserID > 0 && game.discarderUserID != userID {
			discarderPlayerData := game.userIDPlayerDatas[game.discarderUserID]
			prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		}
		countdown := conf.GetCfgTimeout().LandlordDiscard - after
		if countdown > 1 {
			playerData.user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
				ActionDiscardType: game.userIDPlayerDatas[userID].actionDiscardType,
				Position:          game.userIDPlayerDatas[userID].position,
				Countdown:         countdown - 1,
				PrevDiscards:      prevDiscards,
			})
		}
	}

}

func (game *LandlordMatchRoom) rank() {
	/*
	   若遇到总得分相同时，则按照如下规则进行破同分：

	   （1）首先比较最后一副牌的得分，高者名次列前；

	   （2）其次比较获胜（即得分>0）的牌副数，牌副数多者名次列前；

	   （3）再次比较出牌总时间，出牌时间少者名次列前；

	   （4）最后比较报名顺序，报名早者名次列前。
	*/
	sort.Sort(poker.LstPoker(game.gameRoundResult))
	log.Debug("*************************:%v", game.gameRoundResult)
	for key, value := range game.gameRoundResult {
		log.Debug("key:%v******************:position:%v", key+1, value.Position)
		game.userIDPlayerDatas[game.PositionUserIDs[value.Position]].Level = key + 1
	}
}
func (game *LandlordMatchRoom) sendSimpleScore(userId int) {
	result := make([]msg.Result, 0)
	for _, p := range game.userIDPlayerDatas {
		r := msg.Result{
			TotalScore: p.roundResult.Total,
			Position:   p.position,
		}
		result = append(result, r)
	}
	for key, p := range game.userIDPlayerDatas {
		if key == userId {
			p.user.WriteMsg(&msg.S2C_UpdateTotalScore{
				Result: result,
			})
			break
		}
	}
}
func (game *LandlordMatchRoom) sendUpdateScore() {
	result := make([]msg.Result, 0)
	for _, p := range game.userIDPlayerDatas {
		r := msg.Result{
			TotalScore: p.roundResult.Total,
			Position:   p.position,
		}
		result = append(result, r)
	}
	for _, p := range game.userIDPlayerDatas {
		p.user.WriteMsg(&msg.S2C_UpdateTotalScore{
			Result: result,
		})
	}
}
