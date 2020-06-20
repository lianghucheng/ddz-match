package match

import (
	. "ddz/game/player"
	. "ddz/game/room"
	. "ddz/game/values"
	"ddz/msg"
	"errors"

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

	MatchID       string    // 赛事id号
	MatchName     string    // 赛事名称
	MatchDesc     string    // 赛事描述
	MatchType     string    // 赛事类型
	State         int       // 赛事状态
	MaxPlayer     int       // 最大参赛人数
	SignInPlayers []int     // 比赛报名的所有玩家
	Award         []float64 // 赛事奖金
	AwardDesc     string    // 奖励描述
	AwardTitle    []string  // 赛事title
	AwardContent  []string  // 赛事正文
	EnterFee      int64     // 报名费
	Recommend     string    // 赛事推荐文字信息

	AllPlayers   map[int]*User // 比赛剩余玩家对象
	Rooms        []*Room       // 所有比赛房间对象
	IsClosing    bool          // 是否正在关闭的赛事
	CurrentRound int           // 当前轮次
}

func (base *BaseMatch) SignIn(uid int) error {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
	for _, p := range base.SignInPlayers {
		if p == uid {
			log.Debug("already sign %v", uid)
			user.WriteMsg(&msg.S2C_Apply{
				Error: msg.S2C_Error_Match,
			})
			return errors.New("already signUp")
		}
	}
	if base.myMatch != nil {
		if err := base.myMatch.SignIn(uid); err != nil {
			return err
		}
	}
	base.SignInPlayers = append(base.SignInPlayers, uid)
	base.AllPlayers[uid] = user
	UserIDMatch[uid] = base

	user.WriteMsg(&msg.S2C_Apply{
		Error:  0,
		RaceID: base.MatchID,
		Action: 1,
		Count:  len(base.SignInPlayers),
	})
	Broadcast(&msg.S2C_MatchNum{
		MatchId: base.MatchID,
		Count:   len(base.SignInPlayers),
	})
	// 每签到一个玩家检查一次
	base.CheckStart()
	return nil
}

func (base *BaseMatch) SignOut(uid int) error {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return errors.New("unknown user")
	}
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
		base.SignInPlayers = append(base.SignInPlayers[:index])
	} else {
		base.SignInPlayers = append(base.SignInPlayers[:index], base.SignInPlayers[index+1:]...)
	}

	delete(UserIDMatch, uid)
	delete(base.AllPlayers, uid)
	user.WriteMsg(&msg.S2C_Apply{
		Error:  0,
		RaceID: base.MatchID,
		Action: 2,
		Count:  len(base.SignInPlayers),
	})
	Broadcast(&msg.S2C_MatchNum{
		MatchId: base.MatchID,
		Count:   len(base.SignInPlayers),
	})

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
	if base.myMatch != nil {
		base.myMatch.End()
	}

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

func (base *BaseMatch) SendMatchDetail(uid int) {
	if base.myMatch != nil {
		base.myMatch.SendMatchDetail(uid)
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
