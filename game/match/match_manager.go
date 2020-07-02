package match

import (
	"ddz/game"
	"ddz/game/db"
	. "ddz/game/player"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"strconv"
	"strings"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// NewScoreManager 创建一个新的比赛类型
func NewScoreManager(sc *ScoreConfig) {
	sc.GetAwardItem()
	MatchManagerList[sc.MatchID] = sc
	// 如果是倒计时开赛的赛事，一开始直接创建出子比赛
	if sc.StartType >= 2 {
		sc.CreateOneMatch()
	} else {
		Broadcast(&msg.S2C_RaceInfo{
			Races: GetMatchManagerInfo(1).([]msg.RaceInfo),
		})
	}
}

// SignIn 赛事管理报名
func (sc *ScoreConfig) SignIn(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return
	}
	// 当前没有空闲赛事，则创建一个赛事
	if sc.LastMatch == nil || sc.LastMatch.State != Signing {
		log.Debug("check")
		sc.CreateOneMatch()
	}

	if err := sc.LastMatch.SignIn(uid); err != nil {
		return
	}

	// ok
	sc.AllSignInPlayers = append(sc.AllSignInPlayers, uid)
	// 报名满人则清零
	if len(sc.AllSignInPlayers) >= sc.MaxPlayer && sc.StartType == 1 {
		sc.AllSignInPlayers = []int{}
	}

	user.WriteMsg(&msg.S2C_Apply{
		Error:  0,
		RaceID: sc.MatchID,
		Action: 1,
		Count:  len(sc.AllSignInPlayers),
	})

	Broadcast(&msg.S2C_MatchNum{
		MatchId: sc.MatchID,
		Count:   len(sc.AllSignInPlayers),
	})

	// 赛事结束
	if sc.UseMatch >= sc.TotalMatch {
		delete(MatchManagerList, sc.MatchID)
		// 通知客户端
		RaceInfo := GetMatchManagerInfo(1).([]msg.RaceInfo)
		Broadcast(&msg.S2C_RaceInfo{
			Races: RaceInfo,
		})
		sc.EndMatchManager()
	}
}

// SignOut 赛事管理报名
func (sc *ScoreConfig) SignOut(uid int, matchID string) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return
	}

	// 玩家自己退出比赛
	if _, ok := MatchManagerList[matchID]; ok {
		if sc.LastMatch == nil {
			return
		}
		if err := sc.LastMatch.SignOut(uid); err != nil {
			return
		}
	} else if match, ok := MatchList[matchID]; ok { // 系统赛事倒计时时间到了，人数不够，清理玩家
		// log.Debug("player %v kickout", uid)
		// log.Debug("check:%v", match.SignInPlayers)
		if err := match.SignOut(uid); err != nil {
			return
		}
	}
	// ok
	sc.RemoveSignPlayer(uid)
	user.WriteMsg(&msg.S2C_Apply{
		Error:  0,
		RaceID: sc.MatchID,
		Action: 2,
		Count:  len(sc.AllSignInPlayers),
	})
	Broadcast(&msg.S2C_MatchNum{
		MatchId: sc.MatchID,
		Count:   len(sc.AllSignInPlayers),
	})

}

// GetNormalConfig 获取通用配置
func (sc *ScoreConfig) GetNormalConfig() *NormalCofig {
	c := &NormalCofig{}
	utils.StructCopy(c, sc)
	// log.Debug("check sc:%+v", c)
	return c
}

// GetAwardItem 根据list，解析出具体的奖励物品
func (sc *ScoreConfig) GetAwardItem() {
	list := sc.AwardList
	items := strings.Split(list, ",")
	awards := []string{}
	// log.Debug("check items:%v", items)
	for _, s := range items {
		item := strings.Split(s, ":")
		if len(item) < 2 {
			continue
		}
		one := strings.Split(item[1][:len(item[1])-1], `"`)
		if len(one) < 2 {
			continue
		}
		awards = append(awards, one[1])
		// awards = append(awards, item[1][:len(item[1])-1])
	}
	sc.Award = awards
	// log.Debug("check award:%v", sc.Award)
}

// SendMatchDetail 发送赛事详情
func (sc *ScoreConfig) SendMatchDetail(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Debug("unknow user:%v", uid)
		return
	}
	signNumDetail := sc.StartType == 1
	isSign := false
	if m, ok := UserIDMatch[uid]; ok {
		// c := m.Manager.GetNormalConfig()
		log.Debug("check,%v", m.NormalCofig.MatchID)
		log.Debug("check,%v", sc.MatchID)
		if m.NormalCofig.MatchID == sc.MatchID {
			isSign = true
		}
	}
	sTime := sc.StartTime
	if sc.StartType == 2 {
		sTime = sc.ReadyTime - time.Now().Unix()
	}
	data := &msg.S2C_RaceDetail{
		ID:            sc.MatchID,
		Desc:          sc.MatchName,
		AwardDesc:     sc.AwardDesc,
		AwardList:     sc.AwardList,
		MatchType:     sc.MatchType,
		RoundNum:      sc.RoundNum,
		StartTime:     sTime,
		StartType:     sc.StartType,
		ConDes:        sc.MatchDesc,
		SignNumDetail: signNumDetail,
		EnterFee:      float64(sc.EnterFee),
		SignNum:       len(sc.AllSignInPlayers),
		IsSign:        isSign,
	}
	user.WriteMsg(data)
}

// End 赛事結束的逻辑
func (sc *ScoreConfig) End(matchID string) {
	// match, ok := MatchList[matchID]
	// if !ok {
	// 	return
	// }
	// for _, p := range match.SignInPlayers {
	// 	sc.RemoveSignPlayer(p)
	// }
	game.GetSkeleton().Go(func() {
		db.UpdateMatchManager(sc.MatchID, bson.M{"$set": bson.M{"usematch": sc.UseMatch}})
	}, nil)
}

// RemoveSignPlayer 清除签到玩家
func (sc *ScoreConfig) RemoveSignPlayer(uid int) {
	for i, p := range sc.AllSignInPlayers {
		if p != uid {
			continue
		}
		if i == len(sc.AllSignInPlayers)-1 {
			sc.AllSignInPlayers = sc.AllSignInPlayers[:i]
		} else {
			sc.AllSignInPlayers = append(sc.AllSignInPlayers[:i], sc.AllSignInPlayers[i+1:]...)
		}
	}
}

// CreateOneMatch 创建子赛事
func (sc *ScoreConfig) CreateOneMatch() {
	matchID := ""
	if sc.CurrentIDIndex < len(sc.OfficalIDs) {
		matchID = sc.OfficalIDs[sc.CurrentIDIndex]
	} else {
		matchID = sc.MatchID + strconv.FormatInt(time.Now().Unix(), 10)
	}
	// log.Debug("check:%v", matchID)
	nSconfig := &ScoreConfig{}
	utils.StructCopy(nSconfig, sc)
	nSconfig.MatchID = matchID
	newMatch := NewScoreMatch(nSconfig)
	newMatch.NormalCofig = sc.GetNormalConfig()
	newMatch.Manager = sc
	sc.LastMatch = newMatch
	sc.UseMatch++
	sc.CurrentIDIndex++
	if sc.StartType == 2 {
		sc.ReadyTime = time.Now().Unix() + sc.StartTime
	}
	Broadcast(&msg.S2C_RaceInfo{
		Races: GetMatchManagerInfo(1).([]msg.RaceInfo),
	})
}

// EndMatchManager 该类赛事结束
func (sc *ScoreConfig) EndMatchManager() {
	// 修改赛事配置数据
	game.GetSkeleton().Go(func() {
		db.UpdateMatchManager(sc.MatchID, bson.M{"$set": bson.M{"state": Ending}})
	}, nil)
}
