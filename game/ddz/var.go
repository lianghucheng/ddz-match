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

	"gopkg.in/mgo.v2/bson"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
)

var skeleton = game.GetSkeleton()

// 房间状态
const (
	roomIdle = iota // 0 空闲
	roomGame        // 1 游戏中
)

//赛事规则
type LandlordMatchRule struct {
	MatchId    string    // 赛事ID
	MaxPlayers int       // 人数: 2、3
	BaseScore  int       // 底分:
	Tickets    int64     // 需要消耗的点券
	Award      float64   // 奖励金额
	RoundNum   string    // 赛制
	Round      int       // 对打几轮
	MatchType  string    // 赛事类型
	Desc       string    // 赛事名称
	Awards     []float64 // 数组下标对应名次，值对应该名词的奖励，待开发
	GameType   int       //todo:奖金赛还是金币，待开发
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
	user              *User
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
	roundResult       *poker.LandlordPlayerRoundResult

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
	userIDPlayerDatas map[int]*LandlordMatchPlayerData // Key: userID
	cards             []int                            // 洗好的牌
	lastThree         []int                            // 最后三张
	discards          []int                            // 玩家出的牌
	rests             []int                            // 剩余的牌
	gameRoundResult   []poker.LandlordPlayerRoundResult
	dealerUserID      int         // 庄家 userID(庄家第一个叫地主)
	landlordUserID    int         // 地主 userID
	peasantUserIDs    []int       // 农民 userID
	discarderUserID   int         // 最近一次出牌的人 userID
	inits             map[int]int //  发送getallplay命令的次数
	finisherUserID    int         // 上一局出完牌的人 userID(做下一局庄家)
	spring            bool        // 春天

	bidTimer      *timer.Timer
	doubleTimer   *timer.Timer
	discardTimer  *timer.Timer
	prepareTimer  *timer.Timer
	maxscore      int //当前房间的最大叫分
	winnerUserIDs []int
	gameRecords   map[int]*DDZGameRecord
	Match         values.Match // 比赛对象
}

func (game *LandlordMatchRoom) broadcast(msg interface{}, positionUserIDs map[int]int, pos int) {
	for position, userID := range positionUserIDs {
		if position == pos {
			continue
		}
		if playerData, ok := game.userIDPlayerDatas[userID]; ok {
			if playerData.user.State != 1 {
				playerData.user.WriteMsg(msg)
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
	dealerPlayerData := game.userIDPlayerDatas[game.dealerUserID]
	dealerPlayerData.dealer = true
	// 确定闲家(注：闲家的英文单词也为player)
	dealerPos := dealerPlayerData.position
	for i := 1; i < game.rule.MaxPlayers; i++ {
		playerPos := (dealerPos + i) % game.rule.MaxPlayers
		playerUserID := game.PositionUserIDs[playerPos]
		playerPlayerData := game.userIDPlayerDatas[playerUserID]
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
	game.spring = false
	game.winnerUserIDs = []int{}
	game.inits = make(map[int]int)
	game.maxscore = 0
	game.gameRoundResult = make([]poker.LandlordPlayerRoundResult, 0)
}

func (game *LandlordMatchRoom) initplayerData() {
	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
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
		log.Debug("当前总分:%v,玩家赢的次数:%v", playerData.roundResult.Total, playerData.roundResult.Wins)
		playerData.roundResult.Chips = 0
		playerData.count = 0
	}
}

func (game *LandlordMatchRoom) allWaiting() bool {
	count := 0
	for _, userID := range game.PositionUserIDs {
		playerData := game.userIDPlayerDatas[userID]
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
		playerData := game.userIDPlayerDatas[userID]
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
	game.spring = true
	landlordWin := true
	landlordPlayerData := game.userIDPlayerDatas[game.landlordUserID]
	var loserUserIDs []int // 用于连胜任务统计
	if game.landlordUserID == game.winnerUserIDs[0] {
		loserUserIDs = append(loserUserIDs, game.peasantUserIDs...)
		for _, peasantUserID := range game.peasantUserIDs {
			peasantPlayerData := game.userIDPlayerDatas[peasantUserID]
			if len(peasantPlayerData.discards) > 0 {
				game.spring = false
				break
			}
		}

		log.Debug("游戏结束 地主胜利 春天: %v", game.spring)
	} else {
		landlordWin = false
		game.winnerUserIDs = game.peasantUserIDs
		loserUserIDs = append(loserUserIDs, game.landlordUserID)
		if len(landlordPlayerData.discards) > 1 {
			game.spring = false
		}
		log.Debug("游戏结束 农民胜利 春天: %v", game.spring)
	}
	if game.spring {
		for userID, player := range game.userIDPlayerDatas {
			//春天
			if game.landlordUserID == game.winnerUserIDs[0] {
				player.Spring = 1
				player.Dealer *= 2

				game.sendRoomPanel(userID)
				continue
			}
			//反春天
			player.LSpring = 1
			if player.user.BaseData.UserData.UserID == game.landlordUserID {
				player.Xian *= 2
				game.sendRoomPanel(userID)
				continue
			}
			player.Xian *= 2
			game.sendRoomPanel(userID)
		}

	}

	for userID, player := range game.userIDPlayerDatas {
		//公共*庄家*防守
		log.Debug("玩家%v的公共%v庄家%v防守%v", player.user.BaseData.UserData.UserID, player.Public, player.Dealer, player.Xian)
		if landlordWin {
			if player.user.BaseData.UserData.UserID != game.landlordUserID {
				player.roundResult.Chips = -int64(player.Public * player.Dealer * player.Xian)
				game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
			} else {
				player.roundResult.Chips = int64(player.Public * player.Dealer * player.Xian)
				player.wins++
				game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				game.gameRecords[userID].Result[game.count-1].Identity = 1
				game.gameRecords[userID].Result[game.count-1].Event = 1
			}
		} else {
			if player.user.BaseData.UserData.UserID == game.landlordUserID {
				player.roundResult.Chips = -int64(player.Public * player.Dealer * player.Xian)
				game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				game.gameRecords[userID].Result[game.count-1].Identity = 1
			} else {
				player.roundResult.Chips = int64(player.Public * player.Dealer * player.Xian)
				player.wins++
				game.gameRecords[userID].Result[game.count-1].Score = player.roundResult.Chips
				game.gameRecords[userID].Result[game.count-1].Event = 1
			}
		}
	}
}

func (game *LandlordMatchRoom) empty() bool {
	return len(game.PositionUserIDs) == 0
}

// 发送单局结果
func (game *LandlordMatchRoom) SendRoundResult(userID int) {
	if playerData, ok := game.userIDPlayerDatas[userID]; ok {
		roundResults := game.gameRoundResult
		result := poker.ResultLose
		Type := 0
		if utils.InArray(game.winnerUserIDs, userID) {
			result = poker.ResultWin
		}
		if playerData.user.BaseData.UserData.UserID == game.landlordUserID {
			Type = 1
		}
		tempMsg := &msg.S2C_LandlordRoundResult{
			Result:       result,
			Spring:       game.spring,
			RoundResults: roundResults,
			ContinueGame: true,
			Type:         Type,
			Position:     playerData.position,
			Allcount:     game.rule.Round,
			CurrCount:    game.count,
			RankOrder:    playerData.Level,
			Process:      game.GetProcess(),
			Countdown:    conf.GetCfgTimeout().LandlordNextStart,
		}
		playerData.user.WriteMsg(tempMsg)
		for _, value := range roundResults {
			log.Debug("单局结算分数:%v,总分:%v,尾牌得分:%v", value.Chips, value.Total, value.Last)
		}
	}
}

func (game *LandlordMatchRoom) sendRoomPanel(userID int) {
	game.userIDPlayerDatas[userID].sendRoomPanel(game.rule.BaseScore)
}

func (ctx *LandlordMatchPlayerData) sendRoomPanel(baseScore int) {
	ctx.user.WriteMsg(&msg.S2C_RoomPanel{
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

func (game *LandlordMatchRoom) sendMineRoundRank(userID int) {
	playerData := game.userIDPlayerDatas[userID]
	game.matchEndMail(userID, playerData.Level, 0)
	result := poker.ResultLose
	Type := 0
	if utils.InArray(game.winnerUserIDs, userID) {
		result = poker.ResultWin
	}
	if playerData.user.BaseData.UserData.UserID == game.landlordUserID {
		Type = 1
	}

	award := float64(0)
	if playerData.Level < len(game.rule.Awards)+1 {
		award = game.rule.Awards[playerData.Level]
		playerData.user.BaseData.UserData.Fee += utils.Decimal(award * 0.8)
		UpdateUserData(playerData.user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"fee": playerData.user.BaseData.UserData.Fee}})
		playerData.user.WriteMsg(&msg.S2C_UpdateUserAfterTaxAward{
			AfterTaxAward: playerData.user.BaseData.UserData.Fee,
		})
	}
	playerData.user.WriteMsg(&msg.S2C_MineRoundRank{
		Result:    result,
		RankOrder: playerData.Level,
		Award:     utils.Decimal(award * 0.8),
		Spring:    game.spring,
		Type:      Type,
	})
	game.gameRecords[userID].Award = utils.Decimal(award * 0.8)
	game.gameRecords[userID].Level = playerData.Level
	game.gameRecords[userID].Count = game.count
	game.gameRecords[userID].Last = playerData.roundResult.Last
	game.gameRecords[userID].Total = playerData.roundResult.Total
	game.gameRecords[userID].Wins = playerData.wins
	for _, playerData := range game.userIDPlayerDatas {
		r := values.Rank{
			Level:    playerData.Level,
			NickName: playerData.user.BaseData.UserData.Nickname,
			Count:    game.count,
			Total:    playerData.roundResult.Total,
			Last:     playerData.roundResult.Last,
			Wins:     playerData.wins,
			Period:   game.gameRecords[playerData.user.BaseData.UserData.UserID].Period,
			Award:    game.gameRecords[playerData.user.BaseData.UserData.UserID].Award,
			Sort:     playerData.roundResult.Sort,
		}
		game.gameRecords[userID].Rank = append(game.gameRecords[userID].Rank, r)
	}

}

//todo：卡住原因，每个玩家的当场总得分和总时长尚未开发，当场游戏的类型和前多少名的奖励未开发
func (game *LandlordMatchRoom) FlushRank(gametype int, rankType string) {
	cfghall := conf.GetCfgHall()
	switch rankType {
	case cfghall.RankTypeJoinNum:
		for _, v := range game.PositionUserIDs {
			hall.FlushRank(gametype, rankType, v, 0)
		}
	case cfghall.RankTypeWinNum:
		for _, v := range game.winnerUserIDs {
			hall.FlushRank(gametype, rankType, v, 0)
		}
	case cfghall.RankTypeFailNum:
		allUserid := []int{}
		for _, v := range game.PositionUserIDs {
			allUserid = append(allUserid, v)
		}
		for _, v := range utils.Remove(allUserid, game.winnerUserIDs) {
			hall.FlushRank(gametype, rankType, v, 0)
		}
	case cfghall.RankTypeAward:
		for k, v := range game.userIDPlayerDatas {
			if v.Level < len(game.rule.Awards) {
				skeleton.ChanRPCServer.Go("WriteAwardFlowData", &msg.RPC_WriteAwardFlowData{
					Userid:  k,
					Amount:  game.rule.Awards[v.Level],
					Matchid: game.rule.MatchId,
				})
				hall.FlushRank(gametype, rankType, v.user.BaseData.UserData.UserID, game.rule.Awards[v.Level])
			}
		}
	}
}

func (game *LandlordMatchRoom) matchEndMail(userid, order int, award float64) {
	skeleton.ChanRPCServer.Go("SendMatchEndMail", &msg.RPC_SendMatchEndMail{
		Userid:  userid,
		Matchid: game.rule.MatchId,
		Order:   order,
		Award:   award, //playerData.Award,
	})
}

func (game *LandlordMatchRoom) matchInterrupt(userid int, conpon int) {
	skeleton.ChanRPCServer.Go("SendInterruptMail", &msg.RPC_SendInterruptMail{
		Userid:  userid,
		Matchid: game.rule.MatchId,
	})
}
