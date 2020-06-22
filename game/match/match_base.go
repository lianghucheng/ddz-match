package match

import (
	. "ddz/game/player"
	. "ddz/game/room"
	. "ddz/game/values"
	"ddz/msg"
	"errors"
	"strings"

	"github.com/szxby/tools/log"
)

// 赛事种类
const (
	Score = "海选赛"
)

// 赛事状态
const (
	Signing = iota // 报名中
	Playing        // 比赛中
	Ending         // 结算中
)

// BaseMatch 通用的比赛对象
type BaseMatch struct {
	myMatch Match // 不同的赛事

	MatchID       string // 赛事id号
	State         int    // 赛事状态
	MaxPlayer     int    // 最大参赛人数
	SignInPlayers []int  // 比赛报名的所有玩家
	AwardList     string // 赛事奖励列表

	AllPlayers   map[int]*User // 比赛剩余玩家对象
	Rooms        []*Room       // 所有比赛房间对象
	IsClosing    bool          // 是否正在关闭的赛事
	CurrentRound int           // 当前轮次
	Award        []string      // 赛事奖励
	Manager      MatchManager  // 隶属于哪个管理下的赛事
}

func (base *BaseMatch) SignIn(uid int) error {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	if _, ok := UserIDMatch[uid]; ok {
		log.Debug("already sign other %v", uid)
		user.WriteMsg(&msg.S2C_Apply{
			Error: msg.S2C_Error_Match,
		})
		return errors.New("already signUp")
	}
	if base.myMatch != nil {
		if err := base.myMatch.SignIn(uid); err != nil {
			return err
		}
	}
	base.SignInPlayers = append(base.SignInPlayers, uid)
	base.AllPlayers[uid] = user
	UserIDMatch[uid] = base
	// 每签到一个玩家检查一次
	base.CheckStart()
	return nil
}

func (base *BaseMatch) SignOut(uid int) error {
	// user, ok := UserIDUsers[uid]
	// if !ok {
	// 	log.Error("unknow user:%v", uid)
	// 	return errors.New("unknown user")
	// }
	index := -1
	for n, p := range base.SignInPlayers {
		if p == uid {
			index = n
			break
		}
	}
	if index == -1 {
		log.Debug("not sign %v", uid)
		return errors.New("not signUp")
	}

	if base.myMatch != nil {
		if err := base.myMatch.SignOut(uid); err != nil {
			return err
		}
	}
	if index == len(base.SignInPlayers)-1 {
		base.SignInPlayers = base.SignInPlayers[:index]
	} else {
		base.SignInPlayers = append(base.SignInPlayers[:index], base.SignInPlayers[index+1:]...)
	}

	delete(UserIDMatch, uid)
	delete(base.AllPlayers, uid)
	// 清理赛事
	if len(base.SignInPlayers) == 0 && base.IsClosing {
		delete(MatchList, base.MatchID)
	}
	return nil
}

func (base *BaseMatch) CheckStart() {
	if base.myMatch != nil {
		base.myMatch.CheckStart()
	}
}

func (base *BaseMatch) Start() {
	base.State = Playing
	base.CurrentRound++
	if base.myMatch != nil {
		base.myMatch.Start()
	}
}

func (base *BaseMatch) End() {
	base.State = Ending
	for _, uid := range base.SignInPlayers {
		// uid := p.BaseData.UserData.UserID
		game, ok := UserIDRooms[uid]
		if !ok {
			continue
		}
		game.Exit(uid)
		delete(UserIDRooms, uid)
		delete(UserIDMatch, uid)
	}
	base.Manager.End(base.MatchID)

	Broadcast(&msg.S2C_MatchNum{
		MatchId: base.Manager.GetNormalConfig().MatchID,
		Count:   len(base.Manager.GetNormalConfig().AllSignInPlayers),
	})
	if base.myMatch != nil {
		base.myMatch.End()
	}
	delete(MatchList, base.MatchID)
}

func (base *BaseMatch) SplitTable() {
	if base.myMatch != nil {
		base.myMatch.SplitTable()
	}
}

func (base *BaseMatch) RoundOver(roomID string) {
	if base.myMatch != nil {
		base.myMatch.RoundOver(roomID)
	}
}

func (base *BaseMatch) GetRank(uid int) {
	if base.myMatch != nil {
		base.myMatch.GetRank(uid)
	}
}

func (base *BaseMatch) broadcast(msg interface{}) {
	for uid := range base.AllPlayers {
		user, ok := UserIDUsers[uid]
		if !ok {
			continue
		}
		user.WriteMsg(msg)
	}
}

// GetAwardItem 根据list，解析出具体的奖励物品
func (base *BaseMatch) GetAwardItem() {
	list := base.AwardList
	items := strings.Split(list, ",")
	awards := []string{}
	log.Debug("check items:%v", items)
	for _, s := range items {
		item := strings.Split(s, ":")
		if len(item) < 2 {
			continue
		}
		awards = append(awards, item[1])
	}
	base.Award = awards
}
