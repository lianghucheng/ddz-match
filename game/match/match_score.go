package match

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/ddz"
	"ddz/game/hall"
	. "ddz/game/player"
	"ddz/game/poker"
	. "ddz/game/room"
	"ddz/game/values"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// ScoreConfig 配置文件
type ScoreConfig struct {
	// base配置
	MatchID       string   `bson:"matchid"`       // 赛事管理id号 '添加赛事时的必填字段'
	SonMatchID    string   `bson:"sonmatchid"`    // 子赛事id
	State         int      `bson:"state"`         // 赛事状态
	MaxPlayer     int      `bson:"maxplayer"`     // 最大参赛玩家
	SignInPlayers []int    `bson:"signinplayers"` // 比赛报名的所有玩家
	AwardDesc     string   `bson:"awarddesc"`     // 奖励描述
	AwardList     string   `bson:"awardlist"`     // 奖励列表 '添加赛事时的必填字段'
	CreateTime    int64    `bson:"createtime"`    // 比赛创建时间
	MoneyAward    float64  `bson:"moneyaward"`    // 赛事金钱总奖励
	CouponAward   float64  `bson:"couponaward"`   // 赛事点券总奖励
	Award         []string // 具体的赛事奖励

	// score配置
	BaseScore   int64           `bson:"basescore"`   // 基础分数
	StartTime   int64           `bson:"starttime"`   // 比赛开始时间
	LimitPlayer int             `bson:"limitplayer"` // 比赛开始的最少人数 '添加赛事时的必填字段'
	TablePlayer int             `bson:"tableplayer"` // 一桌的游戏人数
	Round       int             `bson:"round"`       // 几局制 '添加赛事时的必填字段'
	Card        int             `bson:"card"`        // 几副制 '添加赛事时的必填字段'
	RoundNum    string          `bson:"roundnum"`    // 赛制制(2局1副)
	StartType   int             `bson:"starttype"`   // 开赛条件(1表示满足三人即可开赛,2表示倒计时多久开赛判断,3表示比赛到点开赛) '添加赛事时的必填字段'
	Eliminate   []int           `bson:"eliminate"`   // 每轮淘汰人数
	Rank        []Rank          `bson:"rank"`        // 整个比赛的总排行
	Record      [][]MatchRecord `bson:"matchrecord"` // 整个比赛的总记录

	// 赛事管理配置
	MatchName  string   `bson:"matchname"`  // 赛事名称 '添加赛事时的必填字段'
	MatchDesc  string   `bson:"matchdesc"`  // 赛事描述
	MatchType  string   `bson:"matchtype"`  // 赛事类型 '添加赛事时的必填字段'
	EnterFee   int64    `bson:"enterfee"`   // 报名费 '添加赛事时的必填字段'
	Recommend  string   `bson:"recommend"`  // 赛事推荐介绍(在赛事列表界面倒计时左侧的文字信息) '添加赛事时的必填字段'
	TotalMatch int      `bson:"totalmatch"` // 后台配置的该种比赛可创建的比赛次数 '添加赛事时的必填字段'
	UseMatch   int      `bson:"usematch"`   // 已使用次数
	OfficalIDs []string `bson:"officalIDs"` // 后台配置的可用比赛id号
	ShelfTime  int64    `bson:"shelftime"`  // 上架时间
	Sort       int      `bson:"sort"`       // 赛事排序 '添加赛事时的必填字段'
	ShowHall   bool     `bson:"showhall"`   // 是否首页展示
	MatchIcon  string   `bson:"matchicon"`  // 赛事图标

	AllSignInPlayers []int        `bson:"-"` // 已报名玩家该种赛事的所有玩家
	CurrentIDIndex   int          `bson:"-"` // 当前赛事取id的序号
	LastMatch        *BaseMatch   `bson:"-"` // 最新分配的一个赛事
	ReadyTime        int64        `bson:"-"` // 比赛开始时间
	StartTimer       *timer.Timer `bson:"-"` // 上架倒计时
}

type sConfig struct {
	BaseScore   int             // 基础分数
	StartTime   int64           // 比赛开始时间
	LimitPlayer int             // 比赛开始的最少人数
	TablePlayer int             // 一桌的游戏人数
	Round       int             // 几局制
	RoundNum    string          // 赛制制(2局1副)
	StartType   int             // 开赛条件（满三人开赛）
	Eliminate   []int           // 每轮淘汰人数
	Rank        []Rank          // 整个比赛的总排行
	Record      [][]MatchRecord // 整个比赛的总记录
	MoneyAward  float64         // 赛事金钱总奖励
	CouponAward float64         // 赛事点券总奖励
}

type scoreMatch struct {
	base          Match
	myConfig      *sConfig
	matchPlayers  []*matchPlayer
	OverRoomCount int                               // 已结束对局并完成上报的房间数
	AllResults    []poker.LandlordPlayerRoundResult // 所有房间打完后，发送给客户端的单轮总结算
}

type matchPlayer struct {
	uid        int
	rank       int
	nickname   string
	totalScore int64
	lastScore  int64
	wins       int
	opTime     int64
	signSort   int
}

// NewScoreMatch 创建一个新的赛事
func NewScoreMatch(c *ScoreConfig) *BaseMatch {
	score := &scoreMatch{}
	score.myConfig = &sConfig{}
	utils.StructCopy(score.myConfig, c)
	score.checkConfig()

	base := &BaseMatch{}
	base.SonMatchID = c.SonMatchID
	base.MaxPlayer = c.MaxPlayer
	base.State = Signing
	base.AwardList = c.AwardList
	base.Award = c.Award
	base.Round = c.Round
	base.AllPlayers = make(map[int]*User)
	base.NormalCofig = c.GetNormalConfig()
	base.CreateTime = time.Now().Unix()

	score.base = base
	base.myMatch = score
	MatchList[base.SonMatchID] = base
	if score.myConfig.StartType == 2 && score.myConfig.StartTime > 0 {
		game.GetSkeleton().AfterFunc(time.Duration(score.myConfig.StartTime)*time.Second, func() {
			base.CheckStart()
		})
	} else if score.myConfig.StartType == 3 && score.myConfig.StartTime > 0 {
		game.GetSkeleton().AfterFunc(time.Duration(score.myConfig.StartTime-time.Now().Unix())*time.Second, func() {
			base.CheckStart()
		})
	}
	return base
}

func (sc *scoreMatch) SignIn(uid int) error {
	base := sc.base.(*BaseMatch)
	c := base.NormalCofig
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	if user.BaseData.UserData.Coupon < c.EnterFee {
		user.WriteMsg(&msg.S2C_Apply{
			Error: msg.S2C_Error_Coupon,
		})
		return errors.New("not enough coupon")
	}
	log.Debug("玩家报名参赛:%v", user.BaseData.UserData.UserID)
	user.BaseData.UserData.Coupon -= c.EnterFee
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
	return nil
}

func (sc *scoreMatch) SignOut(uid int) error {
	base := sc.base.(*BaseMatch)
	c := base.NormalCofig
	user, ok := base.AllPlayers[uid]
	// 玩家不在线
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	_, ok = UserIDUsers[uid]
	user.BaseData.UserData.Coupon += c.EnterFee
	// 玩家不在线
	if !ok {
		UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"Coupon": user.BaseData.UserData.Coupon}})
		return nil
	}
	// hall.UpdateUserCoupon(user, c.EnterFee, user.BaseData.UserData.Coupon-c.EnterFee, user.BaseData.UserData.Coupon, db.MatchOpt, db.MatchSignOut)
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.Coupon(),
	})
	return nil
}

func (sc *scoreMatch) CheckStart() {
	base := sc.base.(*BaseMatch)

	//满人开赛
	if sc.myConfig.StartType == 1 {
		if len(base.SignInPlayers) != base.MaxPlayer {
			return
		}
		base.Start()
	} else if sc.myConfig.StartType >= 2 {
		//赛事开赛人数未达到指定的最小人数(赛事作废,重新开赛)
		if len(base.SignInPlayers) < sc.myConfig.LimitPlayer {
			base.CloseMatch()
		} else {
			base.Start()
		}
		// 重启一个新赛事
		base.Manager.CreateOneMatch()
	}
}

func (sc *scoreMatch) Start() {
	base := sc.base.(*BaseMatch)

	base.broadcast(&msg.S2C_MatchPrepare{
		MatchId: base.SonMatchID,
	})

	// 初始化比赛玩家对象
	for index, uid := range base.SignInPlayers {
		p := base.GetPlayer(uid)
		if p == nil {
			log.Error("unknown player:%v", uid)
			continue
		}
		sc.matchPlayers = append(sc.matchPlayers, &matchPlayer{
			uid:        uid,
			rank:       index + 1,
			nickname:   p.BaseData.UserData.Nickname,
			totalScore: 0,
			lastScore:  0,
			wins:       0,
			opTime:     0,
			signSort:   index + 1,
		})
		matchPlayer := &values.MatchPlayer{
			UID:      uid,
			Rank:     index + 1,
			Nickname: p.BaseData.UserData.Nickname,
			SignSort: index + 1,
			Result:   make([]Result, base.Round),
		}
		p.BaseData.MatchPlayer = matchPlayer
	}

	base.SplitTable()
}

func (sc *scoreMatch) End() {
	base := sc.base.(*BaseMatch)
	// 刷新排行榜
	for _, p := range sc.matchPlayers {
		ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeJoinNum, "", "")
		if p.rank <= len(base.Award) {
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeWinNum, "", "")
			cfg := base.NormalCofig
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeAward, base.Award[p.rank-1], cfg.MatchType)
			hall.WriteFlowData(p.uid, utils.Decimal(values.GetMoneyAward(base.Award[p.rank-1])*0.8), hall.FlowTypeAward, cfg.MatchType, cfg.MatchID, []int{})
		} else {
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeFailNum, "", "")
		}
	}

	// 踢出所有玩家
	for _, p := range base.AllPlayers {
		sc.eliminateOnePlayer(p.BaseData.UserData.UserID)
	}

	// 保存赛事记录
	// 先排序rank
	sc.sortRank()
	sc.sortRecord()
	record := &ScoreConfig{}
	utils.StructCopy(record, base.NormalCofig)
	utils.StructCopy(record, base)
	utils.StructCopy(record, sc.myConfig)
	game.GetSkeleton().Go(func() {
		s := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(s)
		s.DB(db.DB).C("match").Insert(record)
	}, nil)
}

func (sc *scoreMatch) SplitTable() {
	base := sc.base.(*BaseMatch)
	c := base.NormalCofig
	num := len(base.AllPlayers) / sc.myConfig.TablePlayer
	index := 0
	indexs := rand.Perm(len(base.AllPlayers))
	if len(base.Rooms) == 0 {
		rule := &ddz.LandlordMatchRule{
			AllPlayers: len(base.AllPlayers),
			MaxPlayers: sc.myConfig.TablePlayer,
			BaseScore:  sc.myConfig.BaseScore,
			Round:      sc.myConfig.Round,
			MatchId:    base.SonMatchID,
			MatchName:  base.NormalCofig.MatchName,
			Tickets:    c.EnterFee,
			RoundNum:   sc.myConfig.RoundNum,
			Desc:       c.MatchName,
			MatchType:  c.MatchType,
			GameType:   hall.RankGameTypeAward,
			Awards:     base.Award,
			AwardList:  base.AwardList,
			Coupon:     base.NormalCofig.EnterFee,
		}
		for i := 0; i < num; i++ {
			room := InitRoom()
			base.Rooms = append(base.Rooms, room)
			ddzRoom := ddz.LandlordInit(rule)
			ddzRoom.Match = base
			room.Game = ddzRoom
			room.Game.OnInit(room)
		}
	}
	log.Debug("num:%v,rooms:%v", num, len(base.Rooms))
	// 所有玩家先退出原来的房间
	for _, r := range base.Rooms {
		game := r.Game.(*ddz.LandlordMatchRoom)
		for _, playerData := range game.UserIDPlayerDatas {
			log.Debug("kick player:%v", playerData.User.BaseData.UserData.UserID)
			game.Exit(playerData.User.BaseData.UserData.UserID)
		}
		// 房间重置
		game.Reset()
	}
	if num < len(base.Rooms) { // 淘汰玩家后，先拆除房间
		n := len(base.Rooms) - num // 需要拆开的房间数
		base.Rooms = base.Rooms[:len(base.Rooms)-n]
	}
	game.GetSkeleton().AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordNextStart)*time.Millisecond, func() {
		for _, room := range base.Rooms {
			// 随机分配桌子
			for i := 0; i < sc.myConfig.TablePlayer; i++ {
				uid := sc.matchPlayers[indexs[index]].uid
				user := base.AllPlayers[uid]
				if lable := room.Game.Enter(user); lable {
					UserIDRooms[uid] = room.Game
				}
				index++
			}
		}
	})
}

func (sc *scoreMatch) RoundOver(roomID string) {
	base := sc.base.(*BaseMatch)
	sc.OverRoomCount++
	// 比赛未结束
	for n, r := range base.Rooms {
		if r.Number == roomID {
			game := r.Game.(*ddz.LandlordMatchRoom)
			results := []poker.LandlordPlayerRoundResult{}
			// 更新比赛信息
			for _, playerData := range game.UserIDPlayerDatas {
				player := playerData.User.BaseData.MatchPlayer
				for _, p := range sc.matchPlayers {
					if p.uid == player.UID {
						p.lastScore = player.LastScore
						p.totalScore = player.TotalScore
						p.opTime = player.OpTime
						p.wins = player.Wins
						break
					}
				}
				one := poker.LandlordPlayerRoundResult{
					Uid:      playerData.User.BaseData.UserData.UserID,
					Aid:      playerData.User.BaseData.UserData.AccountID,
					Nickname: playerData.User.BaseData.UserData.Nickname,
					Total:    player.TotalScore,
					Last:     player.LastScore,
					Wins:     player.Wins,
					Time:     player.OpTime,
					Sort:     player.SignSort,
				}
				results = append(results, one)
				sc.AllResults = append(sc.AllResults, one)

				// 写入比赛记录
				sc.myConfig.Record[base.CurrentRound-1] = append(sc.myConfig.Record[base.CurrentRound-1], MatchRecord{
					RoundCount: base.CurrentRound,
					CardCount:  base.CurrentCardCount,
					RoomCount:  n + 1,
					UID:        playerData.User.BaseData.UserData.UserID,
					Identity:   player.Result[base.CurrentRound-1].Identity,
					Name:       playerData.User.BaseData.UserData.RealName,
					HandCards:  player.Result[base.CurrentRound-1].HandCards,
					ThreeCards: player.Result[base.CurrentRound-1].ThreeCards,
					Event:      player.Result[base.CurrentRound-1].Event,
					Score:      player.Result[base.CurrentRound-1].Score,
					Multiples:  player.Multiples,
				})
			}
			sort.Sort(poker.LstPoker(results))
			// 发送单局结算信息
			for _, playerData := range game.UserIDPlayerDatas {
				player := playerData.User.BaseData.MatchPlayer
				tempMsg := &msg.S2C_LandlordRoundResult{
					Result:       player.Result[base.CurrentRound-1].Event,
					Spring:       game.Spring,
					RoundResults: results,
					Type:         player.Result[base.CurrentRound-1].Identity,
					CurrCount:    base.CurrentRound,
					Process:      sc.GetProcess(),
					Tables:       len(base.Rooms) - sc.OverRoomCount,
					MatchName:    base.NormalCofig.MatchName,
				}
				playerData.User.WriteMsg(tempMsg)
			}
			break
		}
	}
	// 排序
	sc.sortMatchPlayers()
	// 进入下一局
	sc.NextRound()
}

func (sc *scoreMatch) NextRound() {
	base := sc.base.(*BaseMatch)
	log.Debug("start next round:%v", base.CurrentRound)
	if sc.OverRoomCount < len(base.Rooms) {
		return
	}
	if base.CurrentRound < base.Round {
		// 广播单局总结算
		sort.Sort(poker.LstPoker(sc.AllResults))
		base.broadcast(&msg.S2C_LandlordRoundFinalResult{
			RoundResults: sc.AllResults,
			Countdown:    conf.GetCfgTimeout().LandlordNextStart,
		})

		// 淘汰玩家
		sc.eliminatePlayers()

		// 清理数据
		sc.ClearRoundData()

		// 下局开始，先分桌
		base.CurrentRound++
		base.SplitTable()
	} else {
		base.End()
	}
}

func (sc *scoreMatch) GetRank(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknown user:%v", uid)
		return
	}
	data := []poker.LandlordRankData{}
	for _, p := range sc.matchPlayers {
		data = append(data, poker.LandlordRankData{
			Position: p.rank,
			Nickname: p.nickname,
			Wins:     p.wins,
			Total:    p.totalScore,
			Last:     p.lastScore,
			Time:     p.opTime,
			Sort:     p.signSort,
		})
	}
	user.WriteMsg(&msg.S2C_LandlordMatchRound{
		RoundResults: data,
	})
}

func (sc *scoreMatch) sortMatchPlayers() {
	base := sc.base.(*BaseMatch)
	for i := 0; i < len(base.AllPlayers); i++ {
		for j := i + 1; j < len(base.AllPlayers); j++ {
			// 从大到小排序
			if !rankWay(sc.matchPlayers[i], sc.matchPlayers[j]) {
				// 实际rank为下标+1
				sc.matchPlayers[i].rank = j + 1
				sc.matchPlayers[j].rank = i + 1
				sc.matchPlayers[i], sc.matchPlayers[j] = sc.matchPlayers[j], sc.matchPlayers[i]
			}
		}
		// 同步未被淘汰的玩家的排名信息
		if _, ok := base.AllPlayers[sc.matchPlayers[i].uid]; ok {
			base.AllPlayers[sc.matchPlayers[i].uid].BaseData.MatchPlayer.Rank = sc.matchPlayers[i].rank
		}
	}
}

func (sc *scoreMatch) sortRank() {
	for i := 0; i < len(sc.myConfig.Rank); i++ {
		for j := i + 1; j < len(sc.myConfig.Rank); j++ {
			// 从小到大排序
			if sc.myConfig.Rank[i].Level > sc.myConfig.Rank[j].Level {
				sc.myConfig.Rank[i], sc.myConfig.Rank[j] = sc.myConfig.Rank[j], sc.myConfig.Rank[i]
			}
		}
	}
}

func (sc *scoreMatch) sortRecord() {
	r := sc.myConfig.Record
	for n := range r {
		for i := 0; i < len(r[n]); i++ {
			for j := i + 1; j < len(r[n]); j++ {
				// 从小到大排序
				if r[n][i].RoomCount > r[n][j].RoomCount {
					r[n][i], r[n][j] = r[n][j], r[n][i]
				}
			}
		}
	}
}

// 检查一些配置是否有问题
func (sc *scoreMatch) checkConfig() {
	// 防止配置错误
	if sc.myConfig.TablePlayer < 3 {
		log.Error("error config:%+v", sc.myConfig.TablePlayer)
		sc.myConfig.TablePlayer = 3
	}
	sc.myConfig.Record = make([][]MatchRecord, sc.myConfig.Round)
}

// 淘汰玩家
func (sc *scoreMatch) eliminatePlayers() {
	base := sc.base.(*BaseMatch)
	eliminate := 0 // 淘汰后剩余的玩家数
	log.Debug("start eliminate players:%v,round:%v", sc.myConfig.Eliminate, base.CurrentRound)
	if base.CurrentRound-1 < len(sc.myConfig.Eliminate) {
		eliminate = sc.myConfig.Eliminate[base.CurrentRound-1]
	}
	// eliminate>0代表剩余人数，eliminate<0代表淘汰人数
	if eliminate < 0 {
		eliminate = len(base.AllPlayers) + eliminate
	}
	// 淘汰玩家数过大，导致无法凑成一桌，不淘汰玩家
	if eliminate == 0 || eliminate < sc.myConfig.TablePlayer {
		return
	}
	// 如果剩余玩家无法凑成整数桌，继续淘汰
	last := eliminate % sc.myConfig.TablePlayer
	if last != 0 {
		eliminate -= last
	}
	// 按照排名顺序淘汰玩家
	for n := len(base.AllPlayers) - 1; n > eliminate-1; n-- {
		uid := sc.matchPlayers[n].uid
		sc.eliminateOnePlayer(uid)
	}

	// 广播比赛剩余人数
	// Broadcast(&msg.S2C_MatchNum{
	// 	MatchId: base.MatchID,
	// 	Count:   len(base.Manager.GetNormalConfig().AllSignInPlayers),
	// })
}

// 淘汰指定玩家
func (sc *scoreMatch) eliminateOnePlayer(uid int) {
	log.Debug("eliminate player:%v", uid)
	base := sc.base.(*BaseMatch)
	// 给玩家发送比赛结算总界面
	sc.SendFinalResult(uid)

	// 发奖并记录玩家数据
	sc.awardPlayer(uid)

	if room, ok := UserIDRooms[uid]; ok {
		room.Exit(uid)
	}

	// base.Manager.RemoveSignPlayer(uid)
	delete(UserIDMatch, uid)
	delete(UserIDRooms, uid)
	delete(base.AllPlayers, uid)
}

func (sc *scoreMatch) awardPlayer(uid int) {
	base := sc.base.(*BaseMatch)
	user, ok := base.AllPlayers[uid]
	if !ok {
		log.Error("unknown user:%v", uid)
		return
	}
	player := user.BaseData.MatchPlayer
	var award string
	var moneyAwardCount float64
	if player.Rank-1 < len(base.Award) {
		award = base.Award[player.Rank-1]
		one := strings.Split(award, ",")
		for _, oneAward := range one {
			// 现金奖励
			if values.GetAwardType(oneAward) == values.Money {
				awardAmount := values.ParseAward(oneAward)
				moneyAwardCount += utils.Decimal(awardAmount * 0.8)
				user.BaseData.UserData.Fee += utils.Decimal(awardAmount * 0.8)
				UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"fee": user.BaseData.UserData.Fee}})
				hall.UpdateUserAfterTaxAward(user)
			} else if values.GetAwardType(oneAward) == values.Coupon { // 点券奖励
				awardAmount := values.ParseAward(oneAward)
				user.BaseData.UserData.Coupon += int64(awardAmount)
				UpdateUserData(user.BaseData.UserData.UserID, bson.M{"$set": bson.M{"fee": user.BaseData.UserData.Coupon}})
				hall.UpdateUserCoupon(user, int64(values.ParseAward(oneAward)), user.BaseData.UserData.Coupon-int64(awardAmount),
					user.BaseData.UserData.Coupon, db.MatchOpt, db.MatchAward+fmt.Sprintf("-%v", base.NormalCofig.MatchName))
			}
		}
	}
	// 写入战绩
	record := values.DDZGameRecord{
		UserId:    uid,
		MatchId:   base.SonMatchID,
		MatchType: base.NormalCofig.MatchType,
		Desc:      base.NormalCofig.MatchName,
		Level:     player.Rank,
		Award:     award,
		Count:     base.CurrentRound,
		Total:     player.TotalScore,
		Last:      player.LastScore,
		Wins:      player.Wins,
		Period:    player.OpTime,
		Result:    player.Result[:base.CurrentRound],
		CreateDat: time.Now().Unix(),
	}

	userMatchReview := values.UserMatchReview{}
	wins := 0
	fails := 0
	if len(award) > 0 {
		wins = 1
	} else {
		fails = 1
	}
	update := values.UserMatchReview{
		UID:            uid,
		AccountID:      user.BaseData.UserData.AccountID,
		MatchID:        base.NormalCofig.MatchID,
		MatchType:      base.NormalCofig.MatchType,
		MatchName:      base.NormalCofig.MatchName,
		MatchWins:      wins,
		MatchFails:     fails,
		Coupon:         base.NormalCofig.EnterFee,
		AwardMoney:     int64(moneyAwardCount * 100),
		PersonalProfit: int64(moneyAwardCount*100) - base.NormalCofig.EnterFee,
	}
	var err error
	game.GetSkeleton().Go(
		func() {
			hall.MatchEndPushMail(uid, base.NormalCofig.MatchName, player.Rank, award)
			db.InsertMatchRecord(record)
			userMatchReview, err = db.GetUserMatchReview(uid, base.NormalCofig.MatchType, base.NormalCofig.MatchID)
		}, func() {
			if err != nil {
				log.Error("err:%v", err)
				return
			}
			userMatchReview.MatchWins += update.MatchWins
			userMatchReview.MatchFails += update.MatchFails
			userMatchReview.Coupon += update.Coupon
			userMatchReview.AwardMoney += update.AwardMoney
			userMatchReview.PersonalProfit += update.PersonalProfit
			userMatchReview.MatchTotal = userMatchReview.MatchWins + userMatchReview.MatchFails
			userMatchReview.AverageBatting = userMatchReview.MatchWins / userMatchReview.MatchTotal
			userMatchReview.MatchID = update.MatchID
			userMatchReview.MatchType = update.MatchType
			userMatchReview.UID = update.UID
			userMatchReview.AccountID = update.AccountID
			userMatchReview.MatchName = update.MatchName
			db.UpsertUserMatchReview(bson.M{"uid": userMatchReview.UID, "matchname": userMatchReview.MatchName,
				"matchtype": userMatchReview.MatchType, "matchid": userMatchReview.MatchID}, userMatchReview)
		})

	// 将单个玩家的数据写入rank中
	sc.myConfig.Rank = append(sc.myConfig.Rank, Rank{
		Level:    player.Rank,
		NickName: user.BaseData.UserData.Nickname,
		Count:    base.CurrentRound,
		Total:    player.TotalScore,
		Last:     player.LastScore,
		Wins:     player.Wins,
		Period:   player.OpTime,
		Sort:     player.SignSort,
		Award:    award,
	})

	// 淘汰后清除比赛数据
	user.BaseData.MatchPlayer = nil
}

func rankWay(p1, p2 *matchPlayer) bool {
	if p1.totalScore > p2.totalScore {
		return true
	}
	if p1.totalScore < p2.totalScore {
		return false
	}
	if p1.lastScore > p2.lastScore {
		return true
	}
	if p1.lastScore < p2.lastScore {
		return false
	}
	if p1.wins > p2.wins {
		return true
	}
	if p1.wins < p2.wins {
		return false
	}
	if p1.opTime/100 < p2.opTime/100 { // 只比较到秒的小数点后一位
		return true
	}
	if p1.opTime/100 > p2.opTime/100 { // 只比较到秒的小数点后一位
		return false
	}
	if p1.signSort < p2.signSort {
		return true
	}
	return false
}

// GetProcess 获取进程描述
func (sc *scoreMatch) GetProcess() []string {
	base := sc.base.(*BaseMatch)
	ret := make([]string, base.Round)
	if base.Round == 2 {
		for i := 0; i < base.Round; i++ {
			s := ""
			n := 0
			if i < len(sc.myConfig.Eliminate) {
				n = sc.myConfig.Eliminate[i]
			}
			// 如果n>0代表剩余人数，n<0代表淘汰人数
			if n < 0 {
				n = len(base.AllPlayers) + n
			} else if n == 0 {
				n = len(base.AllPlayers)
			}
			if i == 0 {
				s = fmt.Sprintf("首局:%v人", n)
			} else if i == 1 {
				s = fmt.Sprintf("决赛:%v人", n)
			} else if i == 2 {
				s = fmt.Sprintf("冠军:%v人", 1)
			}
			ret[i] = s
		}
	} else if base.Round == 3 {
		for i := 0; i < base.Round; i++ {
			s := ""
			n := 0
			if i < len(sc.myConfig.Eliminate) {
				n = sc.myConfig.Eliminate[i]
			}
			// 如果n>0代表剩余人数，n<0代表淘汰人数
			if n < 0 {
				n = len(base.AllPlayers) - n
			} else if n == 0 {
				n = len(base.AllPlayers)
			}
			if i == 0 {
				s = fmt.Sprintf("首局:%v人", n)
			} else if i == 1 {
				s = fmt.Sprintf("次局:%v人", n)
			} else if i == 2 {
				s = fmt.Sprintf("决赛:%v人", 1)
			} else if i == 3 {
				s = fmt.Sprintf("冠军:%v人", 1)
			}
			ret[i] = s
		}
	} else {
		for i := 0; i < base.Round; i++ {
			s := ""
			n := 0
			if i < len(sc.myConfig.Eliminate) {
				n = sc.myConfig.Eliminate[i]
			}
			// 如果n>0代表剩余人数，n<0代表淘汰人数
			if n < 0 {
				n = len(base.AllPlayers) - n
			} else if n == 0 {
				n = len(base.AllPlayers)
			}
			if i == base.Round-1 {
				s = fmt.Sprintf("冠军:%v人", n)
			} else {
				s = fmt.Sprintf("第%v局:%v人", i+1, n)
			}
			ret[i] = s
		}
	}
	return ret
}

// ClearRoundData 清除一轮数据
func (sc *scoreMatch) ClearRoundData() {
	sc.OverRoomCount = 0
	sc.AllResults = []poker.LandlordPlayerRoundResult{}
}

// SendRoundResult 给玩家发送单局结算
func (sc *scoreMatch) SendRoundResult(uid int) {
	base := sc.base.(*BaseMatch)
	room, ok := UserIDRooms[uid]
	if !ok {
		log.Debug("unknown player:%v", uid)
		return
	}
	game := room.(*ddz.LandlordMatchRoom)
	user := base.AllPlayers[uid]
	// 发送单局结算信息
	results := []poker.LandlordPlayerRoundResult{}
	for _, playerData := range game.UserIDPlayerDatas {
		player := playerData.User.BaseData.MatchPlayer
		one := poker.LandlordPlayerRoundResult{
			Uid:      playerData.User.BaseData.UserData.UserID,
			Aid:      playerData.User.AcountID(),
			Nickname: playerData.User.BaseData.UserData.Nickname,
			Total:    player.TotalScore,
			Last:     player.LastScore,
			Wins:     player.Wins,
			Time:     player.OpTime,
			Sort:     player.SignSort,
		}
		results = append(results, one)
	}
	sort.Sort(poker.LstPoker(results))
	player := user.BaseData.MatchPlayer
	tempMsg := &msg.S2C_LandlordRoundResult{
		Result:       player.Result[base.CurrentRound-1].Event,
		Spring:       game.Spring,
		RoundResults: results,
		Type:         player.Result[base.CurrentRound-1].Identity,
		CurrCount:    base.CurrentRound,
		Process:      sc.GetProcess(),
		Tables:       len(base.Rooms) - sc.OverRoomCount,
	}
	user.WriteMsg(tempMsg)
}

// SendFinalResult 给玩家发送总结算
func (sc *scoreMatch) SendFinalResult(uid int) {
	base := sc.base.(*BaseMatch)
	user := base.AllPlayers[uid]
	player := user.BaseData.MatchPlayer

	var award []string
	if player.Rank-1 < len(base.Award) {
		for _, one := range strings.Split(base.Award[player.Rank-1], ",") {
			award = append(award, one)
		}
	}
	user.WriteMsg(&msg.S2C_MineRoundRank{
		RankOrder: player.Rank,
		Award:     award,
	})
}
