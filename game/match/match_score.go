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
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"errors"
	"math/rand"
	"time"

	"github.com/szxby/tools/log"
)

// ScoreConfig 配置文件
type ScoreConfig struct {
	// base配置
	MatchID       string   `bson:"matchid"`       // 赛事id号（与赛事管理的matchid不是同一个，共用一个字段）
	State         int      `bson:"state"`         // 赛事状态
	MaxPlayer     int      `bson:"maxplayer"`     // 最大参赛玩家
	SignInPlayers []int    `bson:"signinplayers"` // 比赛报名的所有玩家
	AwardDesc     string   `bson:"awarddesc"`     // 奖励描述
	AwardList     string   `bson:"awardlist"`     // 奖励列表
	Award         []string // 具体的赛事奖励

	// score配置
	BaseScore   int64  `bson:"basescore"`   // 基础分数
	StartTime   int64  `bson:"entertime"`   // 比赛开始时间
	LimitPlayer int    `bson:"limitplayer"` // 比赛开始的最少人数
	TablePlayer int    `bson:"tableplayer"` // 一桌的游戏人数
	Round       int    `bson:"round"`       // 几局制
	RoundNum    string `bson:"roundnum"`    // 赛制制(2局1副)
	StartType   int    `bson:"starttype"`   // 开赛条件(1表示满足三人即可开赛,2表示比赛到点满足多少人条件即可开赛抉择出前几名获取奖励)
	Eliminate   []int  `bson:"eliminate"`   // 每轮淘汰人数

	// 赛事管理配置
	MatchName        string     `bson:"matchname"`        // 赛事名称
	MatchDesc        string     `bson:"matchdesc"`        // 赛事描述
	MatchType        string     `bson:"matchtype"`        // 赛事类型
	MatchRank        []int      `bson:"matchrank"`        // 比賽排序
	EnterFee         int64      `bson:"enterfee"`         // 报名费
	Recommend        string     `bson:"recommend"`        // 赛事推荐介绍(在赛事列表界面倒计时左侧的文字信息)
	TotalMatch       int        `bson:"totalmatch"`       // 后台配置的该种比赛可创建的比赛次数
	OfficalIDs       []string   `bson:"officalIDs"`       // 后台配置的可用比赛id号
	AllSignInPlayers []int      `bson:"allsigninplayers"` // 已报名玩家该种赛事的所有玩家
	Sort             int        `bso:"sort"`              // 赛事排序
	CurrentIDIndex   int        `bson:"-"`                // 当前赛事取id的序号
	LastMatch        *BaseMatch `bson:"-"`                // 最新分配的一个赛事
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
func NewScoreMatch(c *ScoreConfig) *BaseMatch {
	// sConfig := scoreConfig{}
	// if err := json.Unmarshal(c, &sConfig); err != nil {
	// 	log.Error("get config error:%v", err)
	// 	return nil
	// }
	score := &scoreMatch{}
	score.myConfig = &sConfig{}
	// score.BaseScore = c.BaseScore
	// // score.StartTime = sConfig.StartTime
	// score.LimitPlayer = c.LimitPlayer
	// score.TablePlayer = c.TablePlayer
	utils.StructCopy(score.myConfig, c)
	score.checkConfig()

	base := &BaseMatch{}
	base.MatchID = c.MatchID
	// base.MatchName = c.MatchName
	// base.MatchDesc = c.MatchDesc
	// base.MatchType = c.MatchType
	base.MaxPlayer = c.MaxPlayer
	base.State = c.State
	// base.SignInPlayers = c.SignInPlayers
	base.AwardList = c.AwardList
	base.Award = c.Award
	// base.AwardDesc = c.AwardDesc
	// base.EnterFee = c.EnterFee
	// base.Recommend = c.Recommend
	base.AllPlayers = make(map[int]*User)
	// base.GetAwardItem()

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
	c := base.Manager.GetNormalConfig()
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
	hall.UpdateUserCoupon(user, -c.EnterFee, db.MatchSignIn)
	return nil
}

func (sc *scoreMatch) SignOut(uid int) error {
	base := sc.base.(*BaseMatch)
	c := base.Manager.GetNormalConfig()
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	user.BaseData.UserData.Coupon += c.EnterFee
	hall.UpdateUserCoupon(user, c.EnterFee, db.MatchSignOut)
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
			rank:       index + 1,
			nickname:   p.BaseData.UserData.Nickname,
			totalScore: 0,
			lastScore:  0,
			wins:       0,
			opTime:     0,
			signSort:   index + 1,
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

	// 保存赛事记录
	record := &ScoreConfig{}
	utils.StructCopy(record, base.Manager.GetNormalConfig())
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
	c := base.Manager.GetNormalConfig()
	num := len(base.AllPlayers) / sc.myConfig.TablePlayer
	index := 0
	indexs := rand.Perm(len(base.AllPlayers))
	for i := 0; i < num; i++ {
		rule := &ddz.LandlordMatchRule{
			MaxPlayers: base.MaxPlayer,
			BaseScore:  sc.myConfig.BaseScore,
			Round:      sc.myConfig.Round,
			MatchId:    base.MatchID,
			MatchName:  base.Manager.GetNormalConfig().MatchName,
			Tickets:    c.EnterFee,
			RoundNum:   sc.myConfig.RoundNum,
			Desc:       c.MatchName,
			MatchType:  c.MatchType,
			GameType:   hall.RankGameTypeAward,
			Awards:     base.Award,
			AwardList:  base.AwardList,
			Coupon:     int(base.Manager.GetNormalConfig().EnterFee),
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
					// game.SendRoundResult(userID)
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
		// sc.NextRound()
	} else {
		// 比赛结束
		// base.End()
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

func (sc *scoreMatch) sortMatchPlayers() {
	for i := 0; i < len(sc.matchPlayers); i++ {
		for j := i + 1; j < len(sc.matchPlayers); j++ {
			// 从大到小排序
			if !rankWay(sc.matchPlayers[i], sc.matchPlayers[j]) {
				sc.matchPlayers[i].rank = j
				sc.matchPlayers[j].rank = i
				sc.matchPlayers[i], sc.matchPlayers[j] = sc.matchPlayers[j], sc.matchPlayers[i]
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
