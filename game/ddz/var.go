package ddz

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/hall"
	. "ddz/game/player"
	"ddz/game/poker"
	. "ddz/game/room"
	"ddz/game/values"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"fmt"
	"time"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
)

var skeleton = game.GetSkeleton()

// 房间状态
const (
	RoomIdle = iota // 0 空闲
	RoomGame        // 1 游戏中
	Ending          // 结算中
)

//赛事规则
type LandlordMatchRule struct {
	MatchId    string // 赛事ID
	MatchName  string
	AllPlayers int      // 赛事总人数
	MaxPlayers int      // 房间最大人数: 2、3
	BaseScore  int      // 底分:
	Tickets    int64    // 需要消耗的点券
	RoundNum   string   // 赛制
	Round      int      // 对打几轮
	MatchType  string   // 赛事类型
	Desc       string   // 赛事名称
	Awards     []string // 数组下标对应名次，值对应该名次的奖励
	AwardList  string   // 发送给客户端的奖励列表
	GameType   int      //todo:奖金赛还是金币，待开发
	Coupon     int
}

// 玩家状态
const (
	_                       = iota
	landlordReady           // 1 准备
	landlordWaiting         // 2 等待
	landlordActionBid       // 3 前端显示叫地主动作
	landlordActionGrab      // 4 前端显示抢地主动作
	landlordActionDouble    // 5 前端显示加倍动作
	landlordActionShowCards // 6 前端显示明牌动作
	landlordActionDiscard   // 7 前端显示出牌动作
)

// 玩家数据
type LandlordMatchPlayerData struct {
	User              *User
	state             int
	position          int // 用户在桌子上的位置，从 0 开始
	dealer            bool
	wins              int     // 赢的次数
	hands             []int   // 手牌
	discards          [][]int // 打出的牌
	analyzer          *poker.LandlordAnalyzer
	count             int   // 玩家两次不出牌(默认系统出牌(要不起))
	actionDiscardType int   // 出牌动作类型
	actionTimestamp   int64 // 记录动作时间戳
	discardTimeStamp  int64 // 精确到毫秒
	costTimeBydiscard int64 // 精确到毫秒
	// roundResult       *poker.LandlordPlayerRoundResult

	hosted      bool // 是否被托管
	originHands []int
	Sort        int // 报名排序
	Level       int //玩家排名
	score       int //叫分大小
	DealerScore int //叫分
	Ming        int //明牌
	Public      int //公共
	Dealer      int //庄家
	Xian        int //防守方
	Boom        int //炸弹
	Spring      int //春天
	LSpring     int //反春天
}

type LandlordMatchRoom struct {
	count int //房间当前局数
	*Room
	rule              *LandlordMatchRule
	UserIDPlayerDatas map[int]*LandlordMatchPlayerData // Key: userID
	cards             []int                            // 洗好的牌
	lastThree         []int                            // 最后三张
	discards          []int                            // 玩家出的牌
	rests             []int                            // 剩余的牌
	// gameRoundResult   []poker.LandlordPlayerRoundResult
	dealerUserID    int         // 庄家 userID(庄家第一个叫地主)
	landlordUserID  int         // 地主 userID
	peasantUserIDs  []int       // 农民 userID
	discarderUserID int         // 最近一次出牌的人 userID
	inits           map[int]int //  发送getallplay命令的次数
	finisherUserID  int         // 上一局出完牌的人 userID(做下一局庄家)
	Spring          bool        // 春天

	bidTimer      *timer.Timer
	doubleTimer   *timer.Timer
	discardTimer  *timer.Timer
	prepareTimer  *timer.Timer
	maxscore      int //当前房间的最大叫分
	winnerUserIDs []int
	Match         values.Match // 比赛对象
}

func (game *LandlordMatchRoom) broadcast(msg interface{}, positionUserIDs map[int]int, pos int) {
	for position, userID := range positionUserIDs {
		if position == pos {
			continue
		}
		if playerData, ok := game.UserIDPlayerDatas[userID]; ok {
			if playerData.User.State != 1 {
				playerData.User.WriteMsg(msg)
			}
		}
	}
}

func (game *LandlordMatchRoom) prepare() {
	game.initRoom()
	game.initplayerData()
}

func (game *LandlordMatchRoom) initRoom() {
	// 洗牌
	switch game.rule.MaxPlayers {
	case 2:
		game.cards = utils.Shuffle(poker.LandlordAllCards2P)
	case 3:
		game.cards = utils.Shuffle(poker.LandlordAllCards)
	}
	// 确定庄家
	game.dealerUserID = game.PositionUserIDs[0]
	if game.finisherUserID > 0 {
		game.dealerUserID = game.finisherUserID
	}
	game.StartTimestamp = time.Now().Unix()
	game.EachRoundStartTimestamp = game.StartTimestamp
	dealerPlayerData := game.UserIDPlayerDatas[game.dealerUserID]
	dealerPlayerData.dealer = true
	// 确定闲家(注：闲家的英文单词也为player)
	dealerPos := dealerPlayerData.position
	for i := 1; i < game.rule.MaxPlayers; i++ {
		playerPos := (dealerPos + i) % game.rule.MaxPlayers
		playerUserID := game.PositionUserIDs[playerPos]
		playerPlayerData := game.UserIDPlayerDatas[playerUserID]
		playerPlayerData.dealer = false
	}
	game.lastThree = []int{}
	game.discards = []int{}
	// 剩余的牌
	game.rests = game.cards
	game.landlordUserID = -1
	game.peasantUserIDs = []int{}
	game.discarderUserID = -1
	game.finisherUserID = -1
	game.Spring = false
	game.winnerUserIDs = []int{}
	game.inits = make(map[int]int)
	game.maxscore = 0
	// game.gameRoundResult = make([]poker.LandlordPlayerRoundResult, 0)
}

// Reset 重置房间的状态
func (game *LandlordMatchRoom) Reset() {
	game.State = RoomIdle
	game.inits = make(map[int]int)
}

func (game *LandlordMatchRoom) initplayerData() {
	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		playerData.hands = []int{}
		playerData.discards = [][]int{}
		playerData.actionTimestamp = 0
		playerData.hosted = false
		playerData.score = 0
		playerData.DealerScore = 0 //叫分
		playerData.Ming = 0        //明牌
		playerData.Public = 0      //公共
		playerData.Dealer = 0      //庄家
		playerData.Xian = 0        //防守方
		playerData.Boom = 0        //炸弹
		playerData.Spring = 0      //春天
		playerData.LSpring = 0     //反春天
		log.Debug("当前总分:%v,玩家赢的次数:%v", playerData.User.BaseData.MatchPlayer.TotalScore, playerData.User.BaseData.MatchPlayer.Wins)
		// playerData.roundResult.Chips = 0
		playerData.count = 0
	}
}

func (game *LandlordMatchRoom) allWaiting() bool {
	count := 0
	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		if playerData.state == landlordWaiting {
			count++
		}
	}
	if count == game.rule.MaxPlayers {
		return true
	}
	return false
}

func (game *LandlordMatchRoom) showHand() {
	for _, userID := range game.PositionUserIDs {
		playerData := game.UserIDPlayerDatas[userID]
		if len(playerData.hands) > 0 {
			game.broadcast(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
			}, game.PositionUserIDs, -1)
		}
	}
}

// 计算积分
func (game *LandlordMatchRoom) calScore() {
	game.Spring = true
	landlordWin := true
	landlordPlayerData := game.UserIDPlayerDatas[game.landlordUserID]
	var loserUserIDs []int // 用于连胜任务统计
	if game.landlordUserID == game.winnerUserIDs[0] {
		loserUserIDs = append(loserUserIDs, game.peasantUserIDs...)
		for _, peasantUserID := range game.peasantUserIDs {
			peasantPlayerData := game.UserIDPlayerDatas[peasantUserID]
			if len(peasantPlayerData.discards) > 0 {
				game.Spring = false
				break
			}
		}

		log.Debug("游戏结束 地主胜利 春天: %v", game.Spring)
	} else {
		landlordWin = false
		game.winnerUserIDs = game.peasantUserIDs
		loserUserIDs = append(loserUserIDs, game.landlordUserID)
		if len(landlordPlayerData.discards) > 1 {
			game.Spring = false
		}
		log.Debug("游戏结束 农民胜利 春天: %v", game.Spring)
	}
	if game.Spring {
		for userID, player := range game.UserIDPlayerDatas {
			//春天
			if game.landlordUserID == game.winnerUserIDs[0] {
				player.Spring = 1
				player.Dealer *= 2

				game.sendRoomPanel(userID)
				continue
			}
			//反春天
			player.LSpring = 1
			if player.User.BaseData.UserData.UserID == game.landlordUserID {
				player.Xian *= 2
				game.sendRoomPanel(userID)
				continue
			}
			player.Xian *= 2
			game.sendRoomPanel(userID)
		}

	}

	for _, player := range game.UserIDPlayerDatas {
		//公共*庄家*防守
		log.Debug("玩家%v的公共%v庄家%v防守%v", player.User.BaseData.UserData.UserID, player.Public, player.Dealer, player.Xian)
		if landlordWin {
			if player.User.BaseData.UserData.UserID != game.landlordUserID {
				player.User.BaseData.MatchPlayer.LastScore = -int64(player.Public * player.Dealer * player.Xian)
				player.User.BaseData.MatchPlayer.TotalScore += -int64(player.Public * player.Dealer * player.Xian)
				// player.roundResult.Chips = -int64(player.Public * player.Dealer * player.Xian)
				// game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				player.User.BaseData.MatchPlayer.Result[game.count-1].Score = player.User.BaseData.MatchPlayer.LastScore
			} else {
				player.User.BaseData.MatchPlayer.LastScore = int64(player.Public * player.Dealer * player.Xian)
				player.User.BaseData.MatchPlayer.TotalScore += int64(player.Public * player.Dealer * player.Xian)
				// player.roundResult.Chips = int64(player.Public * player.Dealer * player.Xian)
				player.wins++
				player.User.BaseData.MatchPlayer.Wins++
				player.User.BaseData.MatchPlayer.Result[game.count-1].Score = player.User.BaseData.MatchPlayer.LastScore
				player.User.BaseData.MatchPlayer.Result[game.count-1].Identity = 1
				player.User.BaseData.MatchPlayer.Result[game.count-1].Event = 1
				// game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				// game.gameRecords[userID].Result[game.count-1].Identity = 1
				// game.gameRecords[userID].Result[game.count-1].Event = 1
			}
		} else {
			if player.User.BaseData.UserData.UserID == game.landlordUserID {
				player.User.BaseData.MatchPlayer.LastScore = -int64(player.Public * player.Dealer * player.Xian)
				player.User.BaseData.MatchPlayer.TotalScore += -int64(player.Public * player.Dealer * player.Xian)
				// player.roundResult.Chips = -int64(player.Public * player.Dealer * player.Xian)
				player.User.BaseData.MatchPlayer.Result[game.count-1].Score = player.User.BaseData.MatchPlayer.LastScore
				player.User.BaseData.MatchPlayer.Result[game.count-1].Identity = 1
				// game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				// game.gameRecords[userID].Result[game.count-1].Identity = 1
			} else {
				player.User.BaseData.MatchPlayer.LastScore = int64(player.Public * player.Dealer * player.Xian)
				player.User.BaseData.MatchPlayer.TotalScore += int64(player.Public * player.Dealer * player.Xian)
				// player.roundResult.Chips = int64(player.Public * player.Dealer * player.Xian)
				player.wins++
				player.User.BaseData.MatchPlayer.Wins++
				player.User.BaseData.MatchPlayer.Result[game.count-1].Score = player.User.BaseData.MatchPlayer.LastScore
				player.User.BaseData.MatchPlayer.Result[game.count-1].Event = 1
				// game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				// game.gameRecords[userID].Result[game.count-1].Event = 1
			}
		}
		// 记录当前局的所有加倍信息
		// player.user.BaseData.MatchPlayer.Multiples=fmt.Sprintf("春天:%v,炸弹:%v,底分:%v,叫分:%v,明牌:%v,公共:%v,庄家:%v")
	}
}

func (game *LandlordMatchRoom) empty() bool {
	return len(game.PositionUserIDs) == 0
}

// 发送单局结果
// func (game *LandlordMatchRoom) SendRoundResult(userID int) {
// 	if playerData, ok := game.UserIDPlayerDatas[userID]; ok {
// 		roundResults := game.gameRoundResult
// 		result := poker.ResultLose
// 		Type := 0
// 		if utils.InArray(game.winnerUserIDs, userID) {
// 			result = poker.ResultWin
// 		}
// 		if playerData.User.BaseData.UserData.UserID == game.landlordUserID {
// 			Type = 1
// 		}
// 		tempMsg := &msg.S2C_LandlordRoundResult{
// 			Result:       result,
// 			Spring:       game.Spring,
// 			RoundResults: roundResults,
// 			// ContinueGame: true,
// 			Type: Type,
// 			// Position:  playerData.position,
// 			// Allcount:  game.rule.Round,
// 			CurrCount: game.count,
// 			// RankOrder: playerData.Level,
// 			Process: game.GetProcess(),
// 			// Countdown:    conf.GetCfgTimeout().LandlordNextStart,
// 		}
// 		playerData.User.WriteMsg(tempMsg)
// 		for _, value := range roundResults {
// 			log.Debug("单局结算分数:%v,总分:%v,尾牌得分:%v", value.Chips, value.Total, value.Last)
// 		}
// 	}
// }

func (game *LandlordMatchRoom) sendRoomPanel(userID int) {
	game.UserIDPlayerDatas[userID].sendRoomPanel(game.rule.BaseScore)
}

func (ctx *LandlordMatchPlayerData) sendRoomPanel(baseScore int) {
	ctx.User.WriteMsg(&msg.S2C_RoomPanel{
		Spring:      ctx.Spring,
		LSpring:     ctx.LSpring,
		Boom:        ctx.Boom,
		BaseScore:   baseScore,
		DealerScore: ctx.DealerScore,
		Ming:        ctx.Ming,
		Public:      ctx.Public,
		Dealer:      ctx.Dealer,
		Xian:        ctx.Xian,
		Total:       ctx.Public * ctx.Dealer * ctx.Xian,
	})
}

func (game *LandlordMatchRoom) GetProcess() []string {
	if game.rule.Round == 2 {
		return []string{"首局", "决赛", "冠军"}
	} else if game.rule.Round == 3 {
		return []string{"首局", "次局", "决赛", "冠军"}
	} else {
		rt := []string{}
		for i := 0; i < game.rule.Round; i++ {
			rt = append(rt, fmt.Sprintf("第%v局", i+1))
		}
		//todo: 根据具体需求修改
		rt = append(rt, "冠军")
		return rt
	}
}

// func (game *LandlordMatchRoom) sendMineRoundRank(userID int) {
// 	playerData := game.UserIDPlayerDatas[userID]
// 	result := poker.ResultLose
// 	Type := 0
// 	if utils.InArray(game.winnerUserIDs, userID) {
// 		result = poker.ResultWin
// 	}
// 	if playerData.User.BaseData.UserData.UserID == game.landlordUserID {
// 		Type = 1
// 	}

// 	award := float64(0)
// 	if playerData.Level-1 < len(game.rule.Awards) {
// 		// 现金奖励
// 		if values.GetAwardType(game.rule.Awards[playerData.Level-1]) == values.Money {
// 			award = values.ParseAward(game.rule.Awards[playerData.Level-1])
// 			playerData.User.BaseData.UserData.Fee += utils.Decimal(award * 0.8)
// 			UpdateUserData(playerData.User.BaseData.UserData.UserID, bson.M{"$set": bson.M{"fee": playerData.User.BaseData.UserData.Fee}})
// 			playerData.User.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
// 				AfterTaxAward: playerData.User.BaseData.UserData.Fee,
// 			})
// 		} else if values.GetAwardType(game.rule.Awards[playerData.Level-1]) == values.Coupon { // 点券奖励 todo

// 		}
// 	}
// 	playerData.User.WriteMsg(&msg.S2C_MineRoundRank{
// 		Result:    result,
// 		RankOrder: playerData.Level,
// 		Award:     utils.Decimal(award * 0.8),
// 		Spring:    game.Spring,
// 		Type:      Type,
// 	})
// 	game.gameRecords[userID].Award = utils.Decimal(award * 0.8)
// 	game.gameRecords[userID].Level = playerData.Level
// 	game.gameRecords[userID].Count = game.count
// 	game.gameRecords[userID].Last = playerData.roundResult.Last
// 	game.gameRecords[userID].Total = playerData.roundResult.Total
// 	game.gameRecords[userID].Wins = playerData.wins
// 	for _, playerData := range game.UserIDPlayerDatas {
// 		r := values.Rank{
// 			Level:    playerData.Level,
// 			NickName: playerData.User.BaseData.UserData.Nickname,
// 			Count:    game.count,
// 			Total:    playerData.roundResult.Total,
// 			Last:     playerData.roundResult.Last,
// 			Wins:     playerData.wins,
// 			Period:   game.gameRecords[playerData.User.BaseData.UserData.UserID].Period,
// 			Award:    game.gameRecords[playerData.User.BaseData.UserData.UserID].Award,
// 			Sort:     playerData.roundResult.Sort,
// 		}
// 		game.gameRecords[userID].Rank = append(game.gameRecords[userID].Rank, r)
// 		sortRank(game.gameRecords[userID].Rank)
// 	}
// 	game.matchEndMail(userID, playerData.Level, game.gameRecords[userID].Award)
// }

// FlushRank 刷新各种排行榜
func FlushRank(gametype int, uid int, rankType string, award string, matchType string) {
	cfghall := conf.GetCfgHall()
	switch rankType {
	case cfghall.RankTypeJoinNum:
		hall.FlushRank(gametype, rankType, uid, 0)
	case cfghall.RankTypeWinNum:
		hall.FlushRank(gametype, rankType, uid, 0)
	case cfghall.RankTypeFailNum:
		hall.FlushRank(gametype, rankType, uid, 0)
	case cfghall.RankTypeAward:
		log.Debug("【刷新奖金】%v %v, %v, %v, %v", award, len(award), values.GetAwardType(award), values.Money, values.ParseAward(award))
		if len(award) == 0 || values.GetAwardType(award) != values.Money {
			return
		}
		hall.WriteFlowData(uid, values.ParseAward(award), hall.FlowTypeAward, matchType, []int{})
		hall.FlushRank(gametype, rankType, uid, values.ParseAward(award))
	}
}

func (game *LandlordMatchRoom) matchEndMail(userid, order int, award float64) {
	skeleton.ChanRPCServer.Go("SendMatchEndMail", &msg.RPC_SendMatchEndMail{
		Userid:    userid,
		MatchName: game.rule.MatchName,
		Order:     order,
		Award:     award, //playerData.Award,
	})
}

func (game *LandlordMatchRoom) matchInterrupt(userid int, conpon int) {
	skeleton.ChanRPCServer.Go("SendInterruptMail", &msg.RPC_SendInterruptMail{
		Userid:    userid,
		MatchName: game.rule.MatchName,
		Coupon:    game.rule.Coupon,
	})
}

func sortRank(rank []values.Rank) {
	for i := 0; i < len(rank); i++ {
		for j := i + 1; j < len(rank); j++ {
			if rank[i].Level > rank[j].Level {
				rank[i], rank[j] = rank[j], rank[i]
			}
		}
	}
}
