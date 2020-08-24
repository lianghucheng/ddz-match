package match

import (
	"ddz/conf"
	"ddz/config"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/ddz"
	"ddz/game/hall"
	. "ddz/game/player"
	"ddz/game/poker"
	. "ddz/game/room"
	"ddz/game/rpc"
	"ddz/game/values"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// ScoreConfig 配置文件
type ScoreConfig struct {
	// base配置
	MatchSource   int      `bson:"matchsource"`   // 比赛来源,1体总,2自己后台
	MatchLevel    int      `bson:"matchlevel"`    // 体总赛事级别
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
	FragmentAward float64  `bson:"fragmentaward"` // 赛事碎片总奖励
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
	MatchName     string   `bson:"matchname"`     // 赛事名称 '添加赛事时的必填字段'
	MatchDesc     string   `bson:"matchdesc"`     // 赛事描述
	MatchType     string   `bson:"matchtype"`     // 赛事类型 '添加赛事时的必填字段'
	EnterFee      int64    `bson:"enterfee"`      // 报名费 '添加赛事时的必填字段'
	Recommend     string   `bson:"recommend"`     // 赛事推荐介绍(在赛事列表界面倒计时左侧的文字信息) '添加赛事时的必填字段'
	TotalMatch    int      `bson:"totalmatch"`    // 后台配置的该种比赛可创建的比赛次数 '添加赛事时的必填字段'
	UseMatch      int      `bson:"usematch"`      // 已使用次数
	OfficalIDs    []string `bson:"officalIDs"`    // 后台配置的可用比赛id号
	ShelfTime     int64    `bson:"shelftime"`     // 上架时间
	DownShelfTime int64    `bson:"downshelftime"` // 下架时间
	EndTime       int64    `bson:"endtime"`       // 结束时间
	Sort          int      `bson:"sort"`          // 赛事排序 '添加赛事时的必填字段'
	ShowHall      bool     `bson:"showhall"`      // 是否首页展示
	MatchIcon     string   `bson:"matchicon"`     // 赛事图标

	AllSignInPlayers       []int        `bson:"-"` // 已报名玩家该种赛事的所有玩家
	AllPlayingPlayersCount int          `bson:"-"` // 正在参与赛事的玩家总数
	CurrentIDIndex         int          `bson:"-"` // 当前赛事取id的序号
	LastMatch              *BaseMatch   `bson:"-"` // 最新分配的一个赛事
	ReadyTime              int64        `bson:"-"` // 比赛开始时间
	StartTimer             *timer.Timer `bson:"-"` // 上架倒计时
	DownShelfTimer         *timer.Timer `bson:"-"` // 下架倒计时
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
	base                 Match
	myConfig             *sConfig
	matchPlayers         []*matchPlayer
	OverRoomCount        int                               // 已结束对局并完成上报的房间数
	AllResults           []poker.LandlordPlayerRoundResult // 所有房间打完后，发送给客户端的单轮总结算
	AwardResults         SportsCenterAwardResultRet        // 单局发奖结果
	WaitSportCenterTimer *timer.Timer                      // 等待体总回调计时器
	WaitSportCenterCount int                               // 重复调用次数
}

type matchPlayer struct {
	uid            int
	accountID      int
	rank           int
	nickname       string
	totalScore     int64
	lastScore      int64
	wins           int
	opTime         int64
	signSort       int
	awardTimer     *timer.Timer
	result         []Result //牌局详细
	eliminateRound int      // 被淘汰的轮次
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
	// base.CreateTime = time.Now().Unix()
	base.Manager = c

	score.base = base
	base.myMatch = score
	MatchList[base.SonMatchID] = base
	if score.myConfig.StartType == 2 && score.myConfig.StartTime > 0 {
		timer := game.GetSkeleton().AfterFunc(time.Duration(score.myConfig.StartTime)*time.Second, func() {
			base.CheckStart()
		})
		base.Manager.SetStartTimer(timer)
	} else if score.myConfig.StartType == 3 && score.myConfig.StartTime > 0 {
		timer := game.GetSkeleton().AfterFunc(time.Duration(score.myConfig.StartTime-time.Now().Unix())*time.Second, func() {
			base.CheckStart()
		})
		base.Manager.SetStartTimer(timer)
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
		if !user.IsRobot() {
			user.WriteMsg(&msg.S2C_Apply{
				Error: msg.S2C_Error_Coupon,
			})
			return errors.New("not enough coupon")
		}
		log.Debug("机器人加点券")
		user.GetUserData().Coupon += 10 * c.EnterFee
	}
	log.Debug("玩家报名参赛:%v,matchName:%v,matchid:%v,sonid:%v", user.BaseData.UserData.UserID, c.MatchName, c.MatchID, base.SonMatchID)
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
			accountID:  p.BaseData.UserData.AccountID,
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
		if p.rank <= len(sc.matchPlayers)/3 {
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeWinNum, "", "")
			cfg := base.NormalCofig
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeAward, base.Award[p.rank-1], cfg.MatchType)
			hall.WriteMatchAwardRecord(p.uid, cfg.MatchType, cfg.MatchID, cfg.MatchName, base.Award[p.rank-1])
		} else {
			ddz.FlushRank(hall.RankGameTypeAward, p.uid, conf.GetCfgHall().RankTypeFailNum, "", "")
		}
	}

	if base.NormalCofig.MatchSource == MatchSourceSportsCenter {
		game.GetSkeleton().Go(func() {
			ranks := []SportsCenterOneFinalRank{}
			for _, p := range sc.matchPlayers {
				one := SportsCenterOneFinalRank{
					Player_id: strconv.Itoa(p.accountID),
					Ranking:   strconv.Itoa(p.rank),
					// Average_mp_ratio:   " ",
					// Rival_avg_mp_ratio: " ",
					Rank_count: " ",
					Total_time: strconv.FormatInt(p.opTime, 10),
					Status:     "0",
				}
				ranks = append(ranks, one)
			}
			matchID := []byte(base.SonMatchID)
			matchID = append(matchID[:12], matchID[14:]...)
			if _, err := edy_api.FinalRankReport(SportsCenterFinalRankResult{
				Match_id: string(matchID),
				Ranks:    ranks,
			}); err != nil {
				log.Error("err:%v", err)
			} else { // 结果上传完毕
				if _, err := edy_api.RankReportFinish(string(matchID)); err != nil {
					log.Error("err:%v", err)
				}
			}
		}, nil)
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
	// record.EndTime = time.Now().Unix()
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
	game.GetSkeleton().AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordNextStart)*time.Millisecond, func() {
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
		if base.CurrentRound == 1 {
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
		} else {
			// 按排名蛇形分配桌子
			for n, room := range base.Rooms {
				if (n+1)%2 != 0 {
					for i := 0; i < sc.myConfig.TablePlayer; i++ {
						uid := sc.matchPlayers[index+i].uid
						user := base.AllPlayers[uid]
						if lable := room.Game.Enter(user); lable {
							UserIDRooms[uid] = room.Game
						}
					}
				} else {
					for i := sc.myConfig.TablePlayer - 1; i >= 0; i-- {
						uid := sc.matchPlayers[index+i].uid
						user := base.AllPlayers[uid]
						if lable := room.Game.Enter(user); lable {
							UserIDRooms[uid] = room.Game
						}
					}
				}
				index += sc.myConfig.TablePlayer
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
						p.result = player.Result
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

				// 体总数据
				if base.NormalCofig.MatchSource == MatchSourceSportsCenter {
					base.SportsCenterRoundResult = append(base.SportsCenterRoundResult, SportsCenterRoundResult{
						Round_id:             strconv.Itoa(base.CurrentRound),
						Player_id:            strconv.Itoa(playerData.User.BaseData.UserData.AccountID),
						Card_player_id:       strconv.Itoa(player.SignSort),
						Card_numerical_order: strconv.Itoa(base.CurrentCardCount),
						Card_group_id:        "1",
						Card_desk_id:         strconv.Itoa(n + 1),
						Card_score:           strconv.FormatInt(player.LastScore, 10),
						// Mp_score:             " ",
						// Mp_ratio:             " ",
						// Mp_ratio_rank:        " ",
						Card_type:       changeCardsToSportsCenter(playerData.OriginHands),
						Call_score:      strconv.Itoa(playerData.Score),
						Spring:          getSpring(playerData.Spring, playerData.LSpring),
						Raise:           getDouble(playerData.Double),
						Card_hole:       changeCardsToSportsCenter(player.Result[base.CurrentRound-1].ThreeCards),
						Card_rival:      getTablePlayerID(game.UserIDPlayerDatas, player.UID),
						Player_position: strconv.Itoa(playerData.Position),
						Status:          "0",
						Passive:         "0",
					})
				}
			}
			sort.Sort(poker.LstPoker(results))
			// 发送单局结算信息
			notice := "本轮所有牌打完将进入下一轮"
			if base.CurrentRound >= base.Round {
				notice = "赛事结果上报中，请稍后..."
			}
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
					Notice:       notice,
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
	if sc.OverRoomCount < len(base.Rooms) {
		return
	}
	log.Debug("start next round:%v", base.CurrentRound)
	// 发送体总数据
	if base.NormalCofig.MatchSource == MatchSourceSportsCenter {
		ranks := []SportsCenterOneFinalRank{}
		for i := 0; i < len(base.AllPlayers); i++ {
			one := SportsCenterOneFinalRank{
				Player_id: strconv.Itoa(sc.matchPlayers[i].accountID),
				Ranking:   strconv.Itoa(sc.matchPlayers[i].rank),
				// Average_mp_ratio:   " ",
				// Rival_avg_mp_ratio: " ",
				Rank_count: " ",
				Total_time: strconv.FormatInt(sc.matchPlayers[i].opTime, 10),
				Status:     "0",
			}
			ranks = append(ranks, one)
		}
		roundIndex := len(base.AllPlayers)

		// 轮次排名上报
		rankList := []SportsCenterOneRank{}
		for i := 0; i < len(base.AllPlayers); i++ {
			one := SportsCenterOneRank{
				Player_id: strconv.Itoa(sc.matchPlayers[i].accountID),
				Card_rank: strconv.Itoa(sc.matchPlayers[i].rank),
				Status:    "0",
			}
			rankList = append(rankList, one)
		}

		// 是否发奖
		isAward := false
		if base.CurrentRound-1 < len(sc.myConfig.Eliminate) {
			eliminate := sc.myConfig.Eliminate[base.CurrentRound-1]
			if eliminate < len(base.AllPlayers) {
				isAward = true
			}
		} else if base.CurrentRound >= base.Round {
			isAward = true
		}

		matchID := []byte(base.SonMatchID)
		currentRound := []byte(fmt.Sprintf("%02d", base.CurrentRound))
		matchID[len(matchID)-8] = currentRound[0]
		matchID[len(matchID)-7] = currentRound[1]

		sportsCenterRoundResult := []values.SportsCenterRoundResult{}
		for _, v := range base.SportsCenterRoundResult {
			sportsCenterRoundResult = append(sportsCenterRoundResult, values.SportsCenterRoundResult{
				// Match_id: string(matchID),
				// Result_list: base.SportsCenterRoundResult,
				Round_id:             v.Round_id,
				Player_id:            v.Player_id,
				Card_player_id:       v.Card_player_id,
				Card_numerical_order: v.Card_numerical_order,
				Card_group_id:        v.Card_group_id,
				Card_desk_id:         v.Card_desk_id,
				Card_score:           v.Card_score,
				// Mp_score:             v.Mp_score,
				// Mp_ratio:             v.Mp_ratio,
				// Mp_ratio_rank:        v.Mp_ratio_rank,
				Card_type:       v.Card_type,
				Call_score:      v.Call_score,
				Spring:          v.Spring,
				Raise:           v.Raise,
				Card_hole:       v.Card_hole,
				Card_rival:      v.Card_rival,
				Player_position: v.Player_position,
				Status:          v.Status,
				Passive:         v.Passive,
			})
		}

		game.GetSkeleton().Go(func() {
			// 版本1.0.24
			for _, v := range sportsCenterRoundResult {
				// 人人对局结果上报
				if _, err := edy_api.SendMatchResultWithPerson(values.SportsCenterReportPersonal{
					Match_id: string(matchID),
					// Result_list: base.SportsCenterRoundResult,
					Round_id:             v.Round_id,
					Player_id:            v.Player_id,
					Card_player_id:       v.Card_player_id,
					Card_numerical_order: v.Card_numerical_order,
					Card_group_id:        v.Card_group_id,
					Card_desk_id:         v.Card_desk_id,
					Card_score:           v.Card_score,
					// Mp_score:             v.Mp_score,
					// Mp_ratio:             v.Mp_ratio,
					// Mp_ratio_rank:        v.Mp_ratio_rank,
					Card_type:       v.Card_type,
					Call_score:      v.Call_score,
					Spring:          v.Spring,
					Raise:           v.Raise,
					Card_hole:       v.Card_hole,
					Card_rival:      v.Card_rival,
					Player_position: v.Player_position,
					Status:          v.Status,
					Passive:         v.Passive,
				}); err != nil {
					log.Error("err:%v", err)
				}
			}

			// 版本1.0.25
			// edy_api.SendMatchResultWithPerson(values.SportsCenterReportPersonal{
			// 	Match_id:    string(matchID),
			// 	Result_list: base.SportsCenterRoundResult,
			// })

			roundRankResult := values.SportsCenterRankResult{
				Match_id:  string(matchID),
				Round_id:  strconv.Itoa(base.CurrentRound),
				Rank_list: rankList,
			}
			if _, err := edy_api.RoundRankReport(roundRankResult); err != nil {
				log.Error("err:%v", err)
			}

			// 最终排名上报
			if _, err := edy_api.FinalRankReport(SportsCenterFinalRankResult{
				Match_id: string(matchID),
				Ranks:    ranks,
			}); err != nil {
				log.Error("err:%v", err)
			} else { // 结果上传完毕
				if _, err := edy_api.RankReportFinish(string(matchID)); err != nil {
					log.Error("err:%v", err)
				}
			}

			// 最后一轮
			// if base.CurrentRound >= base.Round {
			// 	// 最终排名上报
			// 	if _, err := edy_api.FinalRankReport(SportsCenterFinalRankResult{
			// 		Match_id: string(matchID),
			// 		Ranks:    ranks,
			// 	}); err != nil {
			// 		log.Error("err:%v", err)
			// 	} else { // 结果上传完毕
			// 		if _, err := edy_api.RankReportFinish(string(matchID)); err != nil {
			// 			log.Error("err:%v", err)
			// 		}
			// 	}
			// }

			// 清理数据
			// sc.AwardResults = values.SportsCenterAwardResultRet{}
			if isAward {
				// 拉取体总发奖结果
				sc.getSportsAwardResult(string(matchID), roundIndex)
			}
			// if msg, err := edy_api.AwardResult(string(matchID), 1, roundIndex); err == nil {
			// 	sc.AwardResults = msg
			// }
		}, func() {
			// 等发送完再开下一轮
			sc.onNextRound()
			// if len(sc.AwardResults.Result_list) == 0 {
			// 	return
			// }
			// for _, v := range sc.AwardResults.Result_list {
			// 	// 未发奖
			// 	if v.Status != "2" {
			// 		log.Debug("sports award err:%+v", v)
			// 		continue
			// 	}
			// 	oneUID, err := strconv.Atoi(v.Player_id)
			// 	if err != nil {
			// 		log.Debug("sports award err:%+v", v)
			// 		continue
			// 	}
			// 	base.AwardPlayer(oneUID)
			// }
		})
		return
	}
	sc.onNextRound()
}

func (sc *scoreMatch) getSportsAwardResult(matchID string, num int) {
	base := sc.base.(*BaseMatch)
	// 最多请求5次
	if sc.WaitSportCenterCount > 5 {
		sc.WaitSportCenterCount = 0
		return
	}
	sc.WaitSportCenterTimer = game.GetSkeleton().AfterFunc(1*time.Second, func() {
		game.GetSkeleton().Go(func() {
			// 清理数据
			sc.AwardResults = values.SportsCenterAwardResultRet{}
			// 拉取体总发奖结果
			if msg, err := edy_api.AwardResult(string(matchID), 1, num); err == nil {
				sc.AwardResults = msg
			}
		}, func() {
			if len(sc.AwardResults.Result_list) == 0 {
				sc.WaitSportCenterCount++
				sc.getSportsAwardResult(matchID, num)
				return
			}
			for _, v := range sc.AwardResults.Result_list {
				// 未发奖
				if v.Status != "2" {
					log.Debug("sports award err:%+v", v)
					continue
				}
				oneUID, err := strconv.Atoi(v.Player_id)
				if err != nil {
					log.Debug("sports award err:%+v", v)
					continue
				}
				base.AwardPlayer(oneUID)
			}
		})
	})
}

func (sc *scoreMatch) onNextRound() {
	base := sc.base.(*BaseMatch)
	if base.CurrentRound < base.Round {

		// 淘汰玩家
		sc.eliminatePlayers()

		// 广播单局总结算
		sort.Sort(poker.LstPoker(sc.AllResults))
		base.broadcast(&msg.S2C_LandlordRoundFinalResult{
			RoundResults: sc.AllResults,
			Countdown:    conf.GetCfgTimeout().LandlordNextStart,
		})

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
			AccountID: p.accountID,
			Position:  p.rank,
			Nickname:  p.nickname,
			Wins:      p.wins,
			Total:     p.totalScore,
			Last:      p.lastScore,
			Time:      p.opTime,
			Sort:      p.signSort,
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
	player := sc.getMatchPlayer(uid)
	player.eliminateRound = base.CurrentRound
	// 非体总赛事直接发奖
	if base.NormalCofig.MatchSource != MatchSourceSportsCenter {
		// 给玩家发送比赛结算总界面
		// sc.SendFinalResult(uid)

		// 发奖并记录玩家数据
		base.AwardPlayer(uid)
	} else { // 等待5s体总结果,超时直接结算
		player := sc.getMatchPlayer(uid)
		if player != nil {
			player.awardTimer = game.GetSkeleton().AfterFunc(WaitSportCenterCB*time.Second, func() {
				base.AwardPlayer(uid)
			})
		}
	}

	sc.recordPlayer(uid)

	if room, ok := UserIDRooms[uid]; ok {
		room.Exit(uid)
	}

	// base.Manager.RemoveSignPlayer(uid)
	delete(UserIDMatch, uid)
	delete(UserIDRooms, uid)
	delete(base.AllPlayers, uid)
	// 如果服务器在更新，踢出玩家
	if values.CheckRestart() {
		if user, ok := UserIDUsers[uid]; ok {
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_ServerRestart, Info: values.DefaultRestartConfig})
			user.Close()
			delete(UserIDUsers, uid)
		}
	}
}

// AwardPlayer 发奖
func (sc *scoreMatch) AwardPlayer(uid int) {
	// 发送总结算
	sc.SendFinalResult(uid)

	base := sc.base.(*BaseMatch)
	// user, ok := UserIDUsers[uid]
	// if !ok {
	// 	log.Error("unknown user:%v", uid)
	// 	return
	// }
	// player := user.BaseData.MatchPlayer
	player := sc.getMatchPlayer(uid)
	if player == nil {
		return
	}
	if player.awardTimer != nil {
		player.awardTimer.Stop()
		player.awardTimer = nil
	}
	if sc.WaitSportCenterTimer != nil {
		sc.WaitSportCenterTimer.Stop()
		sc.WaitSportCenterTimer = nil
	}
	status := AwardStatusNormal
	// var moneyAwardCount float64
	if player.rank-1 < len(base.Award) {
		awardStr := base.Award[player.rank-1]
		one := strings.Split(awardStr, ",")
		for _, oneAward := range one {
			log.Debug("award oneAward:%v,type:%v", oneAward, values.GetAwardType(oneAward))
			awardAmount := values.ParseAward(oneAward) * 0.8
			// 现金奖励
			if values.GetAwardType(oneAward) == values.Money {
				// 体总赛事需要对方给了发奖状态才会发奖
				if base.NormalCofig.MatchSource == MatchSourceSportsCenter {
					awardAmount = 0
					for _, v := range sc.AwardResults.Result_list {
						if v.Player_id == strconv.Itoa(player.accountID) {
							if v.Status != "2" {
								log.Error("err award :%+v", v)
								break
							}
							var err error
							awardAmount, err = strconv.ParseFloat(v.Bonous, 64)
							if err != nil {
								log.Error("err award :%+v", v)
							}
							break
						}
					}
				}
				if awardAmount <= 0 {
					status = AwardStatusBonusFail
					continue
				}
				// moneyAwardCount += utils.Decimal(awardAmount * 0.8)
				hall.WriteFlowData(uid, utils.Decimal(awardAmount), hall.FlowTypeAward,
					base.NormalCofig.MatchType, base.NormalCofig.SonMatchID, []int{})
				hall.AddFee(uid, player.accountID, utils.Decimal(awardAmount),
					db.MatchOpt, db.MatchAward+fmt.Sprintf("-%v", base.NormalCofig.MatchName), base.SonMatchID)
			} else if values.GetAwardType(oneAward) == values.Coupon { // 点券奖励
				// awardAmount := values.ParseAward(oneAward)
				hall.AddCoupon(player.uid, player.accountID, int64(awardAmount),
					db.MatchOpt, db.MatchAward+fmt.Sprintf("-%v", base.NormalCofig.MatchName), base.SonMatchID)
			} else if values.GetAwardType(oneAward) == values.Fragment { // 碎片奖励
				hall.AddFragment(uid, player.accountID, int64(awardAmount),
					db.MatchOpt, db.MatchAward+fmt.Sprintf("-%v", base.NormalCofig.MatchName), base.SonMatchID)
			} else if values.GetAwardType(oneAward) == values.RedScore && base.NormalCofig.MatchSource == MatchSourceSportsCenter { // 红分奖励,体总赛事才有分
				// for _, v := range sc.AwardResults.Result_list {
				// 	if v.Player_id == strconv.Itoa(player.accountID) {
				// 		if v.Status != "2" {
				// 			log.Error("err award :%+v", v)
				// 			break
				// 		}
				hall.AddRedScore(uid, player.accountID, awardAmount, db.MatchOpt,
					db.MatchAward+fmt.Sprintf("-%v", base.NormalCofig.MatchName), base.SonMatchID)
				// }
				// }
			}
		}
	}
	award := "道具奖励"
	if sc.myConfig.MoneyAward > 0 {
		award = strconv.FormatFloat(sc.myConfig.MoneyAward, 'f', -1, 64) + "元"
	}
	// 写入战绩
	record := values.DDZGameRecord{
		UserId:    uid,
		MatchId:   base.SonMatchID,
		MatchType: base.NormalCofig.MatchType,
		Desc:      base.NormalCofig.MatchName,
		Level:     player.rank,
		Award:     award,
		Count:     player.eliminateRound,
		Total:     player.totalScore,
		Last:      player.lastScore,
		Wins:      player.wins,
		Period:    player.opTime,
		Result:    player.result[:player.eliminateRound],
		CreateDat: time.Now().Unix(),
		Status:    status,
	}
	// 自己的奖励
	awardStr := ""
	if player.rank-1 < len(base.Award) {
		awardStr = base.Award[player.rank-1]
	}
	game.GetSkeleton().Go(func() {
		if status == AwardStatusNormal {
			hall.MatchEndPushMail(uid, base.NormalCofig.MatchName, player.rank, awardStr)
		} else {
			hall.GamePushMail(uid, "比赛通知", fmt.Sprintf("您在【%v】的参赛结果上报异常，请在战绩中找到对应赛事ID联系客服。谢谢合作", base.NormalCofig.MatchName))
		}
		db.InsertMatchRecord(record)
		rpc.CallActivityServer("DailyWelfareObj.UploadMatchInfo",
			rpc.RPCUploadMatchInfo{AccountID: player.accountID, OptTime: time.Now().Unix() - base.CreateTime}, &rpc.RPCRet{})
	}, nil)
}

// 写入战绩,记录等
func (sc *scoreMatch) recordPlayer(uid int) {
	base := sc.base.(*BaseMatch)
	user, ok := base.AllPlayers[uid]
	if !ok {
		return
	}
	player := user.BaseData.MatchPlayer
	// award := "道具奖励"
	// if sc.myConfig.MoneyAward > 0 {
	// 	award = strconv.FormatFloat(sc.myConfig.MoneyAward, 'f', -1, 64) + "元"
	// }
	// // 写入战绩
	// record := values.DDZGameRecord{
	// 	UserId:    uid,
	// 	MatchId:   base.SonMatchID,
	// 	MatchType: base.NormalCofig.MatchType,
	// 	Desc:      base.NormalCofig.MatchName,
	// 	Level:     player.Rank,
	// 	Award:     award,
	// 	Count:     base.CurrentRound,
	// 	Total:     player.TotalScore,
	// 	Last:      player.LastScore,
	// 	Wins:      player.Wins,
	// 	Period:    player.OpTime,
	// 	Result:    player.Result[:base.CurrentRound],
	// 	CreateDat: time.Now().Unix(),
	// }

	userMatchReview := values.UserMatchReview{}
	wins := 0
	champion := 0
	fails := 0

	var moneyAwardCount float64
	if player.Rank >= len(sc.matchPlayers)/3 {
		wins = 1
	} else {
		fails = 1
	}
	if player.Rank-1 < len(base.Award) {
		moneyAwardCount = values.GetMoneyAward(base.Award[player.Rank-1])
	}

	if player.Rank == 1 {
		champion = 1
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
	gameData := &values.GameData{}
	accountID := user.BaseData.UserData.AccountID

	user.BaseData.UserData.SportCenter.TotalCount++
	user.BaseData.UserData.SportCenter.TotalWin += wins
	user.BaseData.UserData.SportCenter.TotalChampion += champion

	// 赛事总览以及玩家数据记录
	game.GetSkeleton().Go(
		func() {
			// hall.MatchEndPushMail(uid, base.NormalCofig.MatchName, player.Rank, awardStr)
			// db.InsertMatchRecord(record)
			userMatchReview, err = db.GetUserMatchReview(uid, base.NormalCofig.MatchType, base.NormalCofig.MatchID)
			gameData = db.GetUserGameData(uid)
			UpdateUserData(uid, bson.M{"$set": bson.M{"sportcenter": user.BaseData.UserData.SportCenter}})
		}, func() {
			if err == nil {
				// log.Error("err:%v", err)
				// return
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
			}
			if gameData != nil {
				// 无数据初始化
				if gameData.UID <= 0 {
					db.UpsertUserGameData(bson.M{"uid": uid}, values.GameData{
						UID:       uid,
						AccountID: accountID,
						MatchData: &values.MatchData{
							TotalCount: 1,
							WeekCount:  1,
							MonthCount: 1,
							RecordTime: time.Now().Unix(),
						}})

				} else if gameData.MatchData == nil {
					db.UpsertUserGameData(bson.M{"uid": uid}, bson.M{"$set": bson.M{"matchdata": values.MatchData{
						TotalCount: 1,
						WeekCount:  1,
						MonthCount: 1,
						RecordTime: time.Now().Unix(),
					}}})
				} else {
					log.Debug("data:%+v", gameData)
					// 先判断记录时间点
					record := gameData.MatchData.RecordTime
					year, mon, _ := time.Unix(record, 0).Date()
					nYear, nMon, _ := time.Now().Date()
					if year != nYear || (year == nYear && mon != nMon) {
						gameData.MatchData.WeekCount = 1
						gameData.MatchData.MonthCount = 1
					} else {
						gameData.MatchData.MonthCount++
						weekDay := time.Unix(record, 0).Weekday()
						if weekDay == 0 {
							weekDay = 7
						}
						nWeekDay := time.Now().Weekday()
						if nWeekDay == 0 {
							nWeekDay = 7
						}
						// 如果时间差大于一周，清零数据重新统计
						if nWeekDay < weekDay || (time.Now().Unix()-record > 7*24*60*60) {
							gameData.MatchData.WeekCount = 1
						} else {
							gameData.MatchData.WeekCount++
						}
					}
					gameData.MatchData.TotalCount++
					gameData.MatchData.RecordTime = time.Now().Unix()
					db.UpsertUserGameData(bson.M{"uid": uid}, gameData)
				}
			}
		})

	// 自己的奖励
	var awardStr, scoreStr, otherStr string
	if player.Rank-1 < len(base.Award) {
		awardStr = base.Award[player.Rank-1]
		if len(awardStr) > 0 {
			scoreStr, otherStr = values.SplitScoreAward(awardStr)
		}
	}

	// 将单个玩家的数据写入rank中
	sc.myConfig.Rank = append(sc.myConfig.Rank, Rank{
		Level: player.Rank,
		// NickName:   user.BaseData.UserData.Nickname,
		NickName:   strconv.Itoa(user.BaseData.UserData.AccountID),
		Count:      base.CurrentRound,
		Total:      player.TotalScore,
		Last:       player.LastScore,
		Wins:       player.Wins,
		Period:     player.OpTime,
		Sort:       player.SignSort,
		Award:      otherStr,
		ScoreAward: scoreStr,
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
				s = fmt.Sprintf("首轮:%v人", n)
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
				s = fmt.Sprintf("首轮:%v人", n)
			} else if i == 1 {
				s = fmt.Sprintf("次轮:%v人", n)
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
				s = fmt.Sprintf("第%v轮:%v人", i+1, n)
			}
			ret[i] = s
		}
	}
	return ret
}

// ClearRoundData 清除一轮数据
func (sc *scoreMatch) ClearRoundData() {
	base := sc.base.(*BaseMatch)
	sc.OverRoomCount = 0
	sc.AllResults = []poker.LandlordPlayerRoundResult{}
	base.SportsCenterRoundResult = []SportsCenterRoundResult{}
	// sc.AwardResults = SportsCenterAwardResultRet{}
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
	//base := sc.base.(*BaseMatch)
	//user := base.AllPlayers[uid]
	//player := user.BaseData.MatchPlayer
	//
	//var award []string
	//if player.Rank-1 < len(base.Award) {
	//	for _, one := range strings.Split(base.Award[player.Rank-1], ",") {
	//		award = append(award, one)
	//	}
	//}
	//user.WriteMsg(&msg.S2C_MineRoundRank{
	//	RankOrder: player.Rank,
	//	Award:     award,
	//})
	base := sc.base.(*BaseMatch)
	user, ok := UserIDUsers[uid]
	// 玩家不在线
	if !ok {
		return
	}
	cf := config.GetPropBaseConfig
	for _, player := range sc.matchPlayers {
		if player.uid == uid {
			var award []string
			var imgUrl []string
			var awardDatas []map[string]string
			if player.rank-1 < len(base.Award) {
				for _, one := range strings.Split(base.Award[player.rank-1], ",") {
					awardData := make(map[string]string)
					award = append(award, one)
					awardWord := GetAwardType(one)
					propType := AwardWordToPropType[awardWord]
					imgUrl = append(imgUrl, cf(propType).ImgUrl)
					awardData[one] = cf(propType).ImgUrl
					log.Debug("最终结算")
					awardDatas = append(awardDatas, awardData)
				}
			}
			user.WriteMsg(&msg.S2C_MineRoundRank{
				RankOrder: player.rank,
				Award:     awardDatas,
			})
			break
		}
	}
}

// SendMatchInfo 广播赛事信息
func (sc *scoreMatch) SendMatchInfo(uid int) {
	base := sc.base.(*BaseMatch)
	// 广播牌局信息
	eliminate := len(base.AllPlayers)
	if base.CurrentRound-1 < len(sc.myConfig.Eliminate) {
		eliminate = sc.myConfig.Eliminate[base.CurrentRound-1]
	}
	log.Debug("players:%v", base.AllPlayers)
	for _, p := range base.AllPlayers {
		info := msg.S2C_MatchInfo{
			RoundNum:    sc.myConfig.RoundNum,
			Process:     fmt.Sprintf("第%v轮 第1副", base.CurrentRound),
			Level:       fmt.Sprintf("%v/%v", p.BaseData.MatchPlayer.Rank, len(base.AllPlayers)),
			Competition: fmt.Sprintf("前%v晋级", eliminate),
			// AwardList:      base.AwardList,
			AwardList:      getNoneScoreAward(base.Award),
			MatchName:      base.NormalCofig.MatchName,
			Duration:       p.BaseData.MatchPlayer.OpTime,
			WinCnt:         p.BaseData.MatchPlayer.Wins,
			AwardPersonCnt: len(base.Award),
		}
		p.WriteMsg(&info)
	}
}

func (sc *scoreMatch) getMatchPlayer(uid int) *matchPlayer {
	for _, v := range sc.matchPlayers {
		if v.uid == uid {
			return v
		}
	}
	return nil
}
