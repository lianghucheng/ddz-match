package ddz

import (
	"ddz/conf"
	"ddz/game/hall"
	. "ddz/game/player"
	"ddz/game/poker"
	. "ddz/game/room"
	"ddz/msg"
	"fmt"
	"sort"
	"time"

	"github.com/szxby/tools/log"
)

func LandlordInit(rule *LandlordMatchRule) *LandlordMatchRoom {
	roomm := new(LandlordMatchRoom)
	roomm.rule = rule
	return roomm
}
func (game *LandlordMatchRoom) OnInit(room *Room) {
	game.userIDPlayerDatas = make(map[int]*LandlordMatchPlayerData)
	game.Room = room
	game.userIDPlayerDatas = make(map[int]*LandlordMatchPlayerData)
	game.gameRecords = make(map[int]*DDZGameRecord)
	game.inits = make(map[int]int)
}
func (game *LandlordMatchRoom) Play(command interface{}, userId int) {
	switch command.(type) {
	case *msg.C2S_LandlordBid:
		v := command.(*msg.C2S_LandlordBid)
		game.doscore(userId, v.Score)

	case *msg.C2S_LandlordDouble:
		v := command.(*msg.C2S_LandlordDouble)
		game.doDouble(userId, v.Double)

	case *msg.C2S_LandlordDiscard:
		v := command.(*msg.C2S_LandlordDiscard)
		game.doDiscard(userId, v.Cards)

	case *msg.C2S_GetAllPlayers:
		game.GetAllPlayers(game.userIDPlayerDatas[userId].user)

	case *msg.C2S_SystemHost:
		v := command.(*msg.C2S_SystemHost)
		game.doSystemHost(userId, v.Host)

	// case *msg.C2S_LandlordMatchRound:
	// 	{
	// 		game.userIDPlayerDatas[userId].user.WriteMsg(&msg.S2C_LandlordMatchRound{
	// 			RoundResults: game.gameRoundResult,
	// 		})
	// 	}
	default:

	}
}

func (game *LandlordMatchRoom) Enter(user *User) bool {
	userID := user.BaseData.UserData.UserID
	if playerData, ok := game.userIDPlayerDatas[userID]; ok { // 断线重连
		playerData.user = user
		//房间信息返回需要重新规划字段
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_OK,
			Position:   playerData.position,
			BaseScore:  game.rule.BaseScore,
			MaxPlayers: game.rule.MaxPlayers,
		})
		return true
	}
	for pos := 0; pos < game.rule.MaxPlayers; pos++ {
		if _, ok := game.PositionUserIDs[pos]; ok {
			continue
		}
		game.SitDown(user, pos)
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_OK,
			Position:   pos,
			BaseScore:  game.rule.BaseScore,
			MaxPlayers: game.rule.MaxPlayers,
		})
		if len(game.userIDPlayerDatas) == 3 {
			game.prepareTimer = skeleton.AfterFunc(1500*time.Millisecond, func() {
				log.Debug("系统自动发送GetPlayer:开始游戏:%v", 1)
				if game.prepareTimer == nil {
					return
				}
				if game.prepareTimer != nil {
					game.prepareTimer = nil
				}
				for _, p := range game.userIDPlayerDatas {
					game.GetAllPlayers(p.user)
				}
			})
		}
		return true
	}
	return true

}
func (game *LandlordMatchRoom) Exit(userId int) {
	playerData := game.userIDPlayerDatas[userId]
	if playerData == nil {
		return
	}
	playerData.state = 0

	delete(game.PositionUserIDs, playerData.position)
	delete(game.userIDPlayerDatas, userId)
}
func (game *LandlordMatchRoom) SitDown(user *User, pos int) {
	userID := user.BaseData.UserData.UserID
	game.PositionUserIDs[pos] = userID

	playerData := game.userIDPlayerDatas[userID]
	if playerData == nil {
		playerData = new(LandlordMatchPlayerData)
		playerData.user = user
		playerData.position = pos
		playerData.analyzer = new(poker.LandlordAnalyzer)
		playerData.roundResult = new(poker.LandlordPlayerRoundResult)

		game.userIDPlayerDatas[userID] = playerData
	}
}
func (game *LandlordMatchRoom) StandUp(user *User, pos int) {
	delete(game.PositionUserIDs, pos)
	delete(game.userIDPlayerDatas, user.BaseData.UserData.UserID)
}
func (game *LandlordMatchRoom) GetAllPlayers(user *User) {
	_, ok := game.inits[user.BaseData.UserData.UserID]
	game.inits[user.BaseData.UserData.UserID] = 1
	if len(game.inits) == 3 {
		if game.prepareTimer != nil {
			game.prepareTimer.Stop()
			game.prepareTimer = nil
		}
	}
	for pos := 0; pos < game.rule.MaxPlayers; pos++ {
		userID := game.PositionUserIDs[pos]
		if playerData, ok := game.userIDPlayerDatas[userID]; ok {
			playerData.Sort = pos + 1

			msgTemp := &msg.S2C_SitDown{
				Position:   playerData.position,
				AccountID:  playerData.user.BaseData.UserData.AccountID,
				LoginIP:    playerData.user.BaseData.UserData.LoginIP,
				Nickname:   playerData.user.BaseData.UserData.Nickname,
				Headimgurl: playerData.user.BaseData.UserData.Headimgurl,
				Sex:        playerData.user.BaseData.UserData.Sex,
			}
			user.WriteMsg(msgTemp)
		}
	}
	game.sendSimpleScore(user.BaseData.UserData.UserID)
	//断线重连机制
	if game.State == roomIdle && game.count > 0 && game.count < game.rule.Round {

		game.SendRoundResult(user.BaseData.UserData.UserID)
		return
	}
	if game.State == roomIdle && game.count == game.rule.Round {
		game.sendMineRoundRank(user.BaseData.UserData.UserID)
		return
	}
	if game.State == roomGame {
		game.reconnect(user.BaseData.UserData.UserID)
		return
	}
	if !ok {
		if len(game.inits) >= game.rule.MaxPlayers && game.State == roomIdle && game.count == 0 {
			game.StartGame()
			return
		}
	}
	return
}
func (game *LandlordMatchRoom) StartGame() {
	if game.count == 0 {
		for _, userID := range game.PositionUserIDs {

			playerData := game.userIDPlayerDatas[userID]
			game.gameRoundResult = append(game.gameRoundResult, poker.LandlordPlayerRoundResult{
				Uid:      userID,
				Nickname: playerData.user.BaseData.UserData.Nickname,
				Total:    playerData.roundResult.Total,
				Last:     playerData.roundResult.Last,
				Wins:     playerData.wins,
				Time:     playerData.costTimeBydiscard,
				Sort:     playerData.Sort,
				Position: playerData.position,
			})

		}
		sort.Sort(poker.LstPoker(game.gameRoundResult))
		game.GameDDZRecordInit()
	}
	game.State = roomGame
	game.prepare()
	game.count++
	game.broadcast(&msg.S2C_GameStart{}, game.PositionUserIDs, -1)
	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
		info := msg.S2C_MatchInfo{
			RoundNum:    game.rule.RoundNum,
			Process:     fmt.Sprintf("第%v局 第一幅", game.count),
			Level:       fmt.Sprintf("%v/%v", playerData.Level, game.rule.MaxPlayers),
			Competition: "前三晋级",
		}
		playerData.user.WriteMsg(&info)
	}

	// 所有玩家都发十七张牌
	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
		playerData.state = landlordWaiting
		// 手牌有十七张
		playerData.hands = append(playerData.hands, game.rests[:17]...)
		playerData.originHands = playerData.hands
		// 排序
		sort.Sort(sort.Reverse(sort.IntSlice(playerData.hands)))
		log.Debug("userID %v 手牌: %v", userID, poker.ToCardsString(playerData.hands))
		playerData.analyzer.Analyze(playerData.hands)
		// 剩余的牌
		game.rests = game.rests[17:]

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
		game.gameRecords[userID].Result[game.count-1].Count = game.count
		game.gameRecords[userID].Result[game.count-1].HandCards = playerData.hands

	}
	// 庄家叫分
	game.score(game.dealerUserID)
}
func (game *LandlordMatchRoom) EndGame() {

	game.State = roomIdle
	game.showHand()
	game.calScore()

	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
		playerData.roundResult.Total += playerData.roundResult.Chips
		playerData.roundResult.Last = playerData.roundResult.Chips
		playerData.user.WriteMsg(&msg.S2C_ClearAction{})
	}

	game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeWinNum)
	game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeFailNum)

	for _, userID := range game.PositionUserIDs {

		playerData := game.userIDPlayerDatas[userID]
		game.gameRoundResult = append(game.gameRoundResult, poker.LandlordPlayerRoundResult{
			Uid:      userID,
			Nickname: playerData.user.BaseData.UserData.Nickname,
			Total:    playerData.roundResult.Total,
			Last:     playerData.roundResult.Last,
			Wins:     playerData.wins,
			Time:     playerData.costTimeBydiscard,
			Sort:     playerData.Sort,
			Position: playerData.position,
		})
		game.gameRecords[userID].Period += playerData.costTimeBydiscard
	}
	game.EndTimestamp = time.Now().Unix()
	game.rank()
	game.sendUpdateScore()
	// if game.Match != nil {
	// 	game.Match.RoundOver(game.Room.Number)
	// }
	skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordEndPrepare)*time.Millisecond, func() {

		if game.count < game.rule.Round {
			skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordNextStart)*time.Millisecond, func() {
				game.StartGame()
			})
			sort.Sort(poker.LstPoker(game.gameRoundResult))
			for _, userID := range game.PositionUserIDs {

				game.SendRoundResult(userID)
			}

		}

		if game.count == game.rule.Round {
			game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeJoinNum)
			game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeAward)
			for _, userID := range game.PositionUserIDs {
				log.Debug("玩家离开房间:%v", userID)
				game.sendMineRoundRank(userID)
				game.Leave(userID)
			}
			game.GameDDZRecordInsert()
		}

	})
}

// GetRankData 临时处理，获取排行信息
func (game *LandlordMatchRoom) GetRankData(uid int) poker.LandlordRankData {
	playerData := game.userIDPlayerDatas[uid]
	data := poker.LandlordRankData{
		Nickname: playerData.user.BaseData.UserData.Nickname,
		Total:    playerData.roundResult.Total,
		Last:     playerData.roundResult.Last,
		Wins:     playerData.wins,
		Time:     playerData.costTimeBydiscard,
		Sort:     playerData.Sort,
		Position: playerData.position,
	}
	return data
}
