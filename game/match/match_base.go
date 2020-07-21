package match

import (
	"ddz/game/db"
	. "ddz/game/player"
	. "ddz/game/room"
	"ddz/game/values"
	. "ddz/game/values"
	"ddz/msg"
	"errors"
	"fmt"
	"time"

	"github.com/szxby/tools/log"
)

// 赛事种类
const (
	ScoreMatch  = "海选赛"
	MoneyMatch  = "奖金赛"
	DoubleMatch = "复式赛"
	QuickMatch  = "快速赛"
)

// 赛事状态
const (
	Signing = iota // 报名中
	Playing        // 比赛中
	Ending         // 结算中
	Cancel         // 下架赛事
	Delete         // 删除赛事
)

// BaseMatch 通用的比赛对象
type BaseMatch struct {
	myMatch Match // 不同的赛事

	// MatchID       string // 赛事id号
	SonMatchID    string // 子赛事id
	State         int    // 子赛事状态
	MaxPlayer     int    // 最大参赛人数
	SignInPlayers []int  // 比赛报名的所有玩家
	AwardList     string // 赛事奖励列表
	Round         int    // 几局制
	CreateTime    int64  // 比赛创建时间

	AllPlayers       map[int]*User // 比赛剩余玩家对象
	Rooms            []*Room       // 所有比赛房间对象
	IsClosing        bool          // 是否正在关闭的赛事
	CurrentRound     int           // 当前轮次
	CurrentCardCount int           // 当前牌副数
	Award            []string      // 赛事奖励
	Manager          MatchManager  // 隶属于哪个管理下的赛事
	NormalCofig      *NormalCofig  // 一些通用配置
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
	if user.RealName() == "" && !user.IsTest() && !user.IsRobot() {
		log.Debug("no real name. ")
		user.WriteMsg(&msg.S2C_Apply{
			Error: msg.S2C_Error_Realname,
		})
		return errors.New("no real name. ")
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
	if base.NormalCofig.StartType == 1 {
		base.CheckStart()
	}
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
		// base.AllPlayers = base.AllPlayers[:index]
	} else {
		base.SignInPlayers = append(base.SignInPlayers[:index], base.SignInPlayers[index+1:]...)
		// base.AllPlayers = append(base.AllPlayers[:index], base.AllPlayers[index+1:]...)
	}

	delete(UserIDMatch, uid)
	delete(base.AllPlayers, uid)
	// 清理赛事
	if len(base.SignInPlayers) == 0 && base.IsClosing {
		base.Manager.End(base.SonMatchID)
		base.Manager.ClearLastMatch()
		delete(MatchList, base.SonMatchID)
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
	base.CurrentCardCount++
	if base.myMatch != nil {
		base.myMatch.Start()
	}
	// 开赛后清除报名人数
	// c := base.Manager.GetNormalConfig()
	// c.AllSignInPlayers = []int{}
	// base.Manager.SetNormalConfig(c)
	for uid, user := range base.AllPlayers {
		log.Debug("remove uid:%v", uid)
		base.Manager.RemoveSignPlayer(uid)
		// 统计报名费
		// hall.UpdateUserCoupon(user, -base.NormalCofig.EnterFee, user.BaseData.UserData.Coupon+base.NormalCofig.EnterFee,
		// 	user.BaseData.UserData.Coupon, db.MatchOpt, db.MatchSignIn)
		db.InsertItemLog(db.ItemLog{
			UID:        user.BaseData.UserData.AccountID,
			Item:       values.Coupon,
			Amount:     -base.NormalCofig.EnterFee,
			Way:        db.MatchSignIn + fmt.Sprintf(":%v,%v", base.NormalCofig.MatchType, base.NormalCofig.MatchName),
			CreateTime: time.Now().Unix(),
			Before:     user.BaseData.UserData.Coupon + base.NormalCofig.EnterFee,
			After:      user.BaseData.UserData.Coupon,
			OptType:    db.MatchOpt,
			MatchID:    base.SonMatchID,
		})
	}
	BroadcastMatchInfo()
	base.Manager.CheckNewConfig()
}

func (base *BaseMatch) End() {
	base.State = Ending
	// for _, uid := range base.SignInPlayers {
	// 	// uid := p.BaseData.UserData.UserID
	// 	game, ok := UserIDRooms[uid]
	// 	if !ok {
	// 		continue
	// 	}
	// 	game.Exit(uid)
	// 	delete(UserIDRooms, uid)
	// 	delete(UserIDMatch, uid)
	// }
	base.Manager.End(base.SonMatchID)

	// Broadcast(&msg.S2C_MatchNum{
	// 	MatchId: base.Manager.GetNormalConfig().MatchID,
	// 	Count:   len(base.Manager.GetNormalConfig().AllSignInPlayers),
	// })
	if base.myMatch != nil {
		base.myMatch.End()
	}
	delete(MatchList, base.SonMatchID)
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

// GetPlayer 从allplayers中获取
func (base *BaseMatch) GetPlayer(uid int) *User {
	// for _, p := range base.AllPlayers {
	// 	if p.BaseData.UserData.UserID == uid {
	// 		return p
	// 	}
	// }
	return base.AllPlayers[uid]
}

// SendRoundResult 给玩家发送单局结算
func (base *BaseMatch) SendRoundResult(uid int) {
	if base.myMatch != nil {
		base.myMatch.SendRoundResult(uid)
	}
}

// SendFinalResult 给玩家发送单局结算
func (base *BaseMatch) SendFinalResult(uid int) {
	if base.myMatch != nil {
		base.myMatch.SendFinalResult(uid)
	}
}

// CloseMatch 关闭当前赛事
func (base *BaseMatch) CloseMatch() {
	base.IsClosing = true
	// log.Debug("check,%v", base.SignInPlayers)
	// log.Debug("check2:%v", MatchList[base.MatchID].SignInPlayers)
	for _, uid := range base.SignInPlayers {
		base.Manager.SignOut(uid, base.SonMatchID)
	}
}

// GetMatchTypeField 根据赛事类型，返回相应的field name
func (base *BaseMatch) GetMatchTypeField() string {
	switch base.NormalCofig.MatchType {
	case ScoreMatch:
		return "socrematch"
	case MoneyMatch:
		return "moneymatch"
	case DoubleMatch:
		return "doublematch"
	case QuickMatch:
		return "quickmatch"
	default:
		return ""
	}
}
