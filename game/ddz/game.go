package ddz

import (
	. "ddz/game/player"
	"ddz/game/poker"
	. "ddz/game/room"
	"ddz/msg"
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
	game.UserIDPlayerDatas = make(map[int]*LandlordMatchPlayerData)
	game.Room = room
	game.UserIDPlayerDatas = make(map[int]*LandlordMatchPlayerData)
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
		game.GetAllPlayers(game.UserIDPlayerDatas[userId].User)

	case *msg.C2S_SystemHost:
		v := command.(*msg.C2S_SystemHost)
		game.doSystemHost(userId, v.Host)

	// case *msg.C2S_LandlordMatchRound:
	// 	{
	// 		game.UserIDPlayerDatas[userId].User.WriteMsg(&msg.S2C_LandlordMatchRound{
	// 			RoundResults: game.gameRoundResult,
	// 		})
	// 	}
	default:

	}
}

func (game *LandlordMatchRoom) Enter(User *User) bool {
	userID := User.BaseData.UserData.UserID
	if playerData, ok := game.UserIDPlayerDatas[userID]; ok { // 断线重连
		playerData.User = User
		//房间信息返回需要重新规划字段
		User.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_OK,
			Position:   playerData.Position,
			BaseScore:  game.rule.BaseScore,
			MaxPlayers: game.rule.MaxPlayers,
		})
		return true
	}
	log.Debug("player %v enter room:%v", userID, game.Number)
	for pos := 0; pos < game.rule.MaxPlayers; pos++ {
		if _, ok := game.PositionUserIDs[pos]; ok {
			continue
		}
		game.SitDown(User, pos)
		User.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_OK,
			Position:   pos,
			BaseScore:  game.rule.BaseScore,
			MaxPlayers: game.rule.MaxPlayers,
		})
		if len(game.UserIDPlayerDatas) == 3 {
			game.prepareTimer = skeleton.AfterFunc(1500*time.Millisecond, func() {
				log.Debug("系统自动发送GetPlayer:开始游戏:%v", 1)
				if game.prepareTimer == nil {
					return
				}
				if game.prepareTimer != nil {
					game.prepareTimer = nil
				}
				for _, p := range game.UserIDPlayerDatas {
					game.GetAllPlayers(p.User)
				}
			})
		}
		return true
	}
	return true

}
func (game *LandlordMatchRoom) Exit(userId int) {
	playerData := game.UserIDPlayerDatas[userId]
	if playerData == nil {
		return
	}
	playerData.state = 0

	delete(game.PositionUserIDs, playerData.Position)
	delete(game.UserIDPlayerDatas, userId)
}
func (game *LandlordMatchRoom) SitDown(User *User, pos int) {
	userID := User.BaseData.UserData.UserID
	game.PositionUserIDs[pos] = userID

	playerData := game.UserIDPlayerDatas[userID]
	if playerData == nil {
		playerData = new(LandlordMatchPlayerData)
		playerData.User = User
		playerData.Position = pos
		playerData.analyzer = new(poker.LandlordAnalyzer)
		// playerData.roundResult = new(poker.LandlordPlayerRoundResult)

		game.UserIDPlayerDatas[userID] = playerData
	}
}
func (game *LandlordMatchRoom) StandUp(User *User, pos int) {
	delete(game.PositionUserIDs, pos)
	delete(game.UserIDPlayerDatas, User.BaseData.UserData.UserID)
}
func (game *LandlordMatchRoom) GetAllPlayers(User *User) {
	// _, ok := game.inits[User.BaseData.UserData.UserID]
	game.inits[User.BaseData.UserData.UserID] = 1
	if len(game.inits) == 3 {
		if game.prepareTimer != nil {
			game.prepareTimer.Stop()
			game.prepareTimer = nil
		}
	}
	for pos := 0; pos < game.rule.MaxPlayers; pos++ {
		userID := game.PositionUserIDs[pos]
		if playerData, ok := game.UserIDPlayerDatas[userID]; ok {
			playerData.Sort = pos + 1

			msgTemp := &msg.S2C_SitDown{
				Position:   playerData.Position,
				AccountID:  playerData.User.BaseData.UserData.AccountID,
				LoginIP:    playerData.User.BaseData.UserData.LoginIP,
				Nickname:   playerData.User.BaseData.UserData.Nickname,
				Headimgurl: playerData.User.BaseData.UserData.Headimgurl,
				Sex:        playerData.User.BaseData.UserData.Sex,
			}
			User.WriteMsg(msgTemp)
		}
	}
	game.sendSimpleScore(User.BaseData.UserData.UserID)
	//断线重连机制
	if game.State == Ending && game.count > 0 && game.count < game.rule.Round {
		// game.SendRoundResult(User.BaseData.UserData.UserID)
		game.Match.SendRoundResult(User.BaseData.UserData.UserID)
		return
	}
	if game.State == Ending && game.count == game.rule.Round {
		// game.sendMineRoundRank(User.BaseData.UserData.UserID)
		game.Match.SendFinalResult(User.BaseData.UserData.UserID)
		return
	}
	if game.State == RoomGame {
		game.reconnect(User.BaseData.UserData.UserID)
		return
	}
	if len(game.inits) >= game.rule.MaxPlayers && game.State == RoomIdle {
		game.StartGame()
		return
	}
}
func (game *LandlordMatchRoom) StartGame() {
	log.Debug("game %v start", game.Room.Number)
	game.State = RoomGame
	game.prepare()
	game.count++
	game.broadcast(&msg.S2C_GameStart{}, game.PositionUserIDs, -1)

	for _, userID := range game.PositionUserIDs {
		game.Match.SendMatchInfo(userID)
		// 	playerData := game.UserIDPlayerDatas[userID]
		// 	info := msg.S2C_MatchInfo{
		// 		RoundNum:       game.rule.RoundNum,
		// 		Process:        fmt.Sprintf("第%v局 第1副", game.count),
		// 		Level:          fmt.Sprintf("%v/%v", playerData.User.BaseData.MatchPlayer.Rank, game.rule.AllPlayers),
		// 		Competition:    "前3晋级",
		// 		AwardList:      game.rule.AwardList,
		// 		MatchName:      game.rule.MatchName,
		// 		Duration:       playerData.User.BaseData.MatchPlayer.OpTime,
		// 		WinCnt:         playerData.User.BaseData.MatchPlayer.Wins,
		// 		AwardPersonCnt: len(game.rule.Awards),
		// 	}
		// 	playerData.User.WriteMsg(&info)
	}

	// 所有玩家都发十七张牌
	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		playerData.state = landlordWaiting
		// 手牌有十七张
		playerData.hands = append(playerData.hands, game.rests[:17]...)
		playerData.OriginHands = playerData.hands
		// 排序
		sort.Sort(sort.Reverse(sort.IntSlice(playerData.hands)))
		log.Debug("userID %v 手牌: %v", userID, poker.ToCardsString(playerData.hands))
		playerData.analyzer.Analyze(playerData.hands)
		// 剩余的牌
		game.rests = game.rests[17:]

		if playerData, ok := game.UserIDPlayerDatas[userID]; ok {
			playerData.User.WriteMsg(&msg.S2C_UpdatePokerHands{
				Position:      playerData.Position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
			})
		}

		game.broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.Position,
			Hands:         []int{},
			NumberOfHands: len(playerData.hands),
		}, game.PositionUserIDs, playerData.Position)
		// game.gameRecords[userID].Result[game.count-1].Count = game.count
		// game.gameRecords[userID].Result[game.count-1].HandCards = playerData.hands
		// // 目前只有1副牌,todo..
		// game.gameRecords[userID].Result[game.count-1].CardCount = 1
		playerData.User.BaseData.MatchPlayer.Result[game.count-1].Count = game.count
		playerData.User.BaseData.MatchPlayer.Result[game.count-1].HandCards = playerData.hands
		playerData.User.BaseData.MatchPlayer.Result[game.count-1].CardCount = 1
	}
	// 庄家叫分
	game.score(game.dealerUserID)
}
func (game *LandlordMatchRoom) EndGame() {

	game.State = Ending
	game.showHand()
	game.calScore()

	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		// playerData.roundResult.Total += playerData.roundResult.Chips
		// playerData.roundResult.Last = playerData.roundResult.Chips
		playerData.User.WriteMsg(&msg.S2C_ClearAction{})
	}

	// 记录玩家操作时间
	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		// game.gameRoundResult = append(game.gameRoundResult, poker.LandlordPlayerRoundResult{
		// 	Uid:      userID,
		// 	Nickname: playerData.User.BaseData.UserData.Nickname,
		// 	Total:    playerData.User.BaseData.MatchPlayer.TotalScore,
		// 	Last:     playerData.User.BaseData.MatchPlayer.LastScore,
		// 	Wins:     playerData.wins,
		// 	Time:     playerData.costTimeBydiscard,
		// 	Sort:     playerData.Sort,
		// 	Position: playerData.position,
		// })
		// game.gameRecords[userID].Period += playerData.costTimeBydiscard
		playerData.User.BaseData.MatchPlayer.OpTime += playerData.costTimeBydiscard
		playerData.User.BaseData.MatchPlayer.OneOpTime = playerData.costTimeBydiscard
	}
	game.EndTimestamp = time.Now().Unix()
	// game.rank()
	game.sendUpdateScore()

	if game.Match != nil {
		game.Match.RoundOver(game.Room.Number)
	}

	// skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordEndPrepare)*time.Millisecond, func() {

	// 	if game.count < game.rule.Round {
	// 		skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordNextStart)*time.Millisecond, func() {
	// 			game.StartGame()
	// 		})
	// 		sort.Sort(poker.LstPoker(game.gameRoundResult))
	// 		for _, userID := range game.PositionUserIDs {
	// 			game.SendRoundResult(userID)
	// 		}

	// 	}

	// 	if game.count == game.rule.Round {
	// 		game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeJoinNum)
	// 		game.FlushRank(hall.RankGameTypeAward, conf.GetCfgHall().RankTypeAward)
	// 		for _, userID := range game.PositionUserIDs {
	// 			log.Debug("玩家离开房间:%v", userID)
	// 			game.sendMineRoundRank(userID)
	// 			// game.Leave(userID)
	// 		}
	// 		game.GameDDZRecordInsert()
	// 		if game.Match != nil {
	// 			game.Match.End()
	// 		}
	// 	}

	// })
}

// GetRankData 临时处理，获取排行信息
// func (game *LandlordMatchRoom) GetRankData(uid int) poker.LandlordRankData {
// 	playerData := game.UserIDPlayerDatas[uid]
// 	data := poker.LandlordRankData{
// 		Nickname: playerData.User.BaseData.UserData.Nickname,
// 		Total:    playerData.roundResult.Total,
// 		Last:     playerData.roundResult.Last,
// 		Wins:     playerData.wins,
// 		Time:     playerData.costTimeBydiscard,
// 		Sort:     playerData.Sort,
// 		Position: playerData.position,
// 	}
// 	return data
// }
