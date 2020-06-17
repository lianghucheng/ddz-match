package match

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/ddz"
	"ddz/game/hall"
	"ddz/game/poker"
	. "ddz/game/room"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"errors"
	"math/rand"
	"time"

	"github.com/name5566/leaf/log"
)

type scoreConfig struct {
	MatchID       string    `bson:"matchid"`       // 赛事id号
	MatchName     string    `bson:"matchname"`     // 赛事名称
	MatchDesc     string    `bson:"matchdesc"`     // 赛事描述
	MatchType     string    `bson:"matchtype"`     // 赛事类型
	State         int       `bson:"state"`         // 赛事状态
	SignInPlayers []int     `bson:"signinplayers"` // 已报名玩家
	MaxPlayer     int       `bson:"maxplayer"`     // 最大参赛玩家
	Award         []float64 `bson:"award"`         // 赛事奖金
	AwardDesc     string    `bson:"awarddesc"`     // 奖励描述
	AwardTitle    []string  `bson:"awardtitle"`    // 赛事title
	AwardContent  []string  `bson:"awardcontent"`  // 赛事正文
	EnterFee      int64     `bson:"enterfee"`      // 报名费

	BaseScore   int64  `bson:"basescore"`   // 基础分数
	StartTime   int64  `bson:"entertime"`   // 比赛开始时间
	LimitPlayer int    `bson:"limitplayer"` // 比赛开始的最少人数
	TablePlayer int    `bson:"tableplayer"` // 一桌的游戏人数
	Round       int    `bson:"round"`       // 几局制
	RoundNum    string `bson:"roundnum"`    // 赛制制(2局1副)
	StartType   int    `bson:"starttype"`   // 开赛条件(1表示满足三人即可开赛,2表示比赛到点满足多少人条件即可开赛抉择出前几名获取奖励)
	Eliminate   []int  `bson:"eliminate"`   // 每轮淘汰人数
}

type sConfig struct {
	BaseScore   int    // 基础分数
	StartTime   int64  // 比赛开始时间
	LimitPlayer int    // 比赛开始的最少人数
	TablePlayer int    // 一桌的游戏人数
	Round       int    // 几局制
	RoundNum    string // 赛制制(2局1副)
	StartType   int    // 开赛条件（满三人开赛）
	Eliminate   []int  // 每轮淘汰人数
}

type scoreMatch struct {
	base          Match
	myConfig      *sConfig
	matchPlayers  []*matchPlayer
	OverRoomCount int
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
func NewScoreMatch(c *scoreConfig) Match {
	// sConfig := scoreConfig{}
	// if err := json.Unmarshal(c, &sConfig); err != nil {
	// 	log.Error("get config error:%v", err)
	// 	return nil
	// }
	score := &scoreMatch{}
	// score.BaseScore = c.BaseScore
	// // score.StartTime = sConfig.StartTime
	// score.LimitPlayer = c.LimitPlayer
	// score.TablePlayer = c.TablePlayer
	utils.StructCopy(score.myConfig, c)

	base := &BaseMatch{}
	base.MatchID = c.MatchID
	base.MatchName = c.MatchName
	base.MatchDesc = c.MatchDesc
	base.MatchType = c.MatchType
	base.MaxPlayer = c.MaxPlayer
	base.State = c.State
	base.SignInPlayers = c.SignInPlayers
	base.Award = c.Award
	base.AwardDesc = c.AwardDesc
	base.AwardTitle = c.AwardTitle
	base.AwardContent = c.AwardContent
	base.EnterFee = c.EnterFee

	score.base = base
	base.myMatch = score
	MatchList[base.MatchID] = base
	if score.myConfig.StartType == 2 && time.Now().Unix() < score.myConfig.StartTime {
		game.GetSkeleton().AfterFunc(time.Duration(time.Unix(score.myConfig.StartTime, 0).Sub(time.Now()))*time.Second, func() {
			base.CheckStart()
		})
	}
	return base
}

func (sc *scoreMatch) SignIn(uid int) error {
	base := sc.base.(*BaseMatch)
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	if user.BaseData.UserData.Coupon < base.EnterFee {
		user.WriteMsg(&msg.S2C_Apply{
			Error: msg.S2C_Error_Coupon,
		})
		return errors.New("not enough coupon")
	}
	log.Debug("玩家报名参赛:%v", user.BaseData.UserData.UserID)
	user.BaseData.UserData.Coupon -= base.EnterFee
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.BaseData.UserData.Coupon,
	})
	return nil
}

func (sc *scoreMatch) SignOut(uid int) error {
	base := sc.base.(*BaseMatch)
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	user.BaseData.UserData.Coupon += base.EnterFee
	user.WriteMsg(&msg.S2C_UpdateUserCoupon{
		Coupon: user.BaseData.UserData.Coupon,
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
	} else if sc.myConfig.StartType == 2 {
		//赛事开赛人数未达到指定的最小人数(赛事作废,重新开赛)
		if len(base.SignInPlayers) < sc.myConfig.LimitPlayer {
			base.IsClosing = true
			for _, uid := range base.SignInPlayers {
				base.SignOut(uid)
			}
		}
	}
}

func (sc *scoreMatch) Start() {
	base := sc.base.(*BaseMatch)

	base.broadcast(&msg.S2C_MatchPrepare{
		MatchId: base.MatchID,
	})

	// 初始化比赛玩家对象
	for index, uid := range base.SignInPlayers {
		p, ok := base.AllPlayers[uid]
		if !ok {
			log.Error("unknown player:%v", uid)
			continue
		}
		sc.matchPlayers = append(sc.matchPlayers, &matchPlayer{
			uid:        uid,
			rank:       index,
			nickname:   p.BaseData.UserData.Nickname,
			totalScore: 0,
			lastScore:  0,
			wins:       0,
			opTime:     0,
			signSort:   index,
		})
	}

	base.SplitTable()
}

func (sc *scoreMatch) End() {
	base := sc.base.(*BaseMatch)
	for _, r := range base.Rooms {
		game := r.Game.(*ddz.LandlordMatchRoom)
		for _, userID := range game.PositionUserIDs {
			game.SendRoundResult(userID)
		}
	}
}

func (sc *scoreMatch) SplitTable() {
	base := sc.base.(*BaseMatch)
	num := len(base.AllPlayers) / sc.myConfig.TablePlayer
	index := 0
	indexs := rand.Perm(len(base.AllPlayers))
	for i := 0; i < num; i++ {
		rule := &ddz.LandlordMatchRule{
			MaxPlayers: base.MaxPlayer,
			BaseScore:  sc.myConfig.BaseScore,
			Round:      sc.myConfig.Round,
			MatchId:    base.MatchID,
			Tickets:    base.EnterFee,
			RoundNum:   sc.myConfig.RoundNum,
			Desc:       base.MatchName,
			MatchType:  base.MatchType,
			GameType:   hall.RankGameTypeAward,
			Awards:     base.Award,
		}

		room := InitRoom()
		base.Rooms = append(base.Rooms, room)
		ddzRoom := ddz.LandlordInit(rule)
		ddzRoom.Match = base
		room.Game = ddzRoom
		room.Game.OnInit(room)

		// 随机分配桌子
		game.GetSkeleton().AfterFunc(time.Duration(conf.GetCfgTimeout().LandloadMatchPrepare)*time.Millisecond, func() {
			for i := 0; i < sc.myConfig.TablePlayer; i++ {
				uid := base.SignInPlayers[indexs[index]]
				if lable := room.Game.Enter(base.AllPlayers[uid]); lable {
					UserIDRooms[uid] = room.Game
				}
				index++
			}
		})
	}
}

func (sc *scoreMatch) RoundOver(roomID string) {
	base := sc.base.(*BaseMatch)
	sc.OverRoomCount++
	// 比赛未结束
	if base.CurrentRound < sc.myConfig.Round {
		for _, r := range base.Rooms {
			if r.Number == roomID {
				game := r.Game.(*ddz.LandlordMatchRoom)
				// 先发送单局结束面板
				for _, userID := range game.PositionUserIDs {
					game.SendRoundResult(userID)
					data := game.GetRankData(userID)
					for _, p := range sc.matchPlayers {
						if p.uid == userID {
							p.lastScore = data.Last
							p.totalScore = data.Total
							p.opTime = data.Time
							p.wins = data.Wins
							break
						}
					}
				}
			}
			// 排序
			sc.sortMatchPlayers()
		}
		// 进入下一局
		sc.NextRound()
	} else {
		// 比赛结束
		base.End()
	}
}

func (sc *scoreMatch) NextRound() {
	base := sc.base.(*BaseMatch)
	if sc.OverRoomCount < len(base.Rooms) {
		return
	}
	// todo:淘汰玩家
	// todo:下局开始，先分桌
	base.CurrentRound++
	game.GetSkeleton().AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordNextStart)*time.Millisecond, func() {
		for _, r := range base.Rooms {
			r.Game.(*ddz.LandlordMatchRoom).StartGame()
		}
	})
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

func (sc *scoreMatch) SendMatchDetail(uid int) {
	base := sc.base.(*BaseMatch)
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Debug("unknow user:%v", uid)
		return
	}
	signNumDetail := sc.myConfig.StartType == 1
	isSign := false
	if _, ok := UserIDMatch[uid]; ok {
		isSign = true
	}
	data := &msg.S2C_RaceDetail{
		ID:            base.MatchID,
		Desc:          base.MatchName,
		AwardDesc:     base.AwardDesc,
		AwardTitle:    base.AwardTitle,
		AwardContent:  base.AwardContent,
		MatchType:     base.MatchType,
		RoundNum:      sc.myConfig.RoundNum,
		EnterTime:     time.Unix(sc.myConfig.StartTime, 0).Format("2006-01-02 15:04:05"),
		ConDes:        base.MatchDesc,
		SignNumDetail: signNumDetail,
		EnterFee:      float64(base.EnterFee) / 10,
		SignNum:       len(base.SignInPlayers),
		IsSign:        isSign,
	}
	user.WriteMsg(data)
}

// func (sc *scoreMatch) ReStart() {
// 	newConfig := scoreConfig{MatchID: time.Now().Unix()}
// 	base := sc.base.(*BaseMatch)
// 	newConfig.MatchType = base.MatchType
// 	newConfig.State = base.State
// 	newConfig.SignInPlayers = []int{}
// 	newConfig.BaseScore = sc.sConfig.BaseScore
// 	newConfig.LimitPlayer = sc.sConfig.LimitPlayer
// 	// c, _ := json.Marshal(newConfig)
// 	NewScoreMatch(newConfig)
// }

// func (sc *scoreMatch) copyConfig() scoreConfig {
// 	newConfig := scoreConfig{}
// 	newConfig.BaseScore = sc.sConfig.BaseScore
// 	newConfig.LimitPlayer = sc.sConfig.LimitPlayer
// 	base := sc.base.(*BaseMatch)
// 	newConfig.MatchID = base.MatchID
// 	newConfig.MatchType = base.MatchType
// 	newConfig.State = base.State
// 	newConfig.SignInPlayers = base.SignInPlayers
// 	return newConfig
// }
func (sc *scoreMatch) sortMatchPlayers() {
	for i := 0; i < len(sc.matchPlayers); i++ {
		p1 := sc.matchPlayers[i]
		for j := i + 1; j < len(sc.matchPlayers); j++ {
			p2 := sc.matchPlayers[j]
			// 从大到小排序
			if !rankWay(p1, p2) {
				sc.matchPlayers[i], sc.matchPlayers[j] = sc.matchPlayers[j], sc.matchPlayers[i]
				sc.matchPlayers[i].rank = j
				sc.matchPlayers[j].rank = i
			}
		}
	}
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
	if p1.opTime < p2.opTime {
		return true
	}
	if p2.opTime > p2.opTime {
		return false
	}
	if p1.signSort < p2.signSort {
		return true
	}
	return false
}
