package match

import (
	. "ddz/game/player"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"strconv"
	"strings"
	"time"

	"github.com/szxby/tools/log"
)

// NewScoreManager 创建一个新的比赛类型
func NewScoreManager(sc *ScoreConfig) {
	sc.GetAwardItem()
	MatchManagerList[sc.MatchID] = sc
}

// SignIn 赛事管理报名
func (sc *ScoreConfig) SignIn(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return
	}
	if sc.LastMatch == nil || sc.LastMatch.State != Signing {
		// 当前没有空闲赛事，则创建一个赛事
		matchID := ""
		if sc.CurrentIDIndex < len(sc.OfficalIDs) {
			matchID = sc.OfficalIDs[sc.CurrentIDIndex]
		} else {
			matchID = strconv.FormatInt(time.Now().Unix(), 10)
		}
		nSconfig := &ScoreConfig{}
		utils.StructCopy(nSconfig, sc)
		nSconfig.MatchID = matchID
		newMatch := NewScoreMatch(nSconfig)
		newMatch.Manager = sc
		sc.LastMatch = newMatch
		sc.TotalMatch--
		sc.CurrentIDIndex++
	}

	if err := sc.LastMatch.SignIn(uid); err != nil {
		return
	}

	// ok
	tp := sc.TablePlayer
	if tp < 3 {
		tp = 3
	}
	sc.AllSignInPlayers = append(sc.AllSignInPlayers, uid)

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
	if len(sc.AllSignInPlayers)%tp == 0 && sc.TotalMatch <= 0 {
		delete(MatchManagerList, sc.MatchID)
		// 通知客户端
		RaceInfo := GetMatchManagerInfo()
		Broadcast(&msg.S2C_RaceInfo{
			Races: RaceInfo,
		})
	}
}

// SignOut 赛事管理报名
func (sc *ScoreConfig) SignOut(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return
	}

	if sc.LastMatch == nil {
		return
	}
	if err := sc.LastMatch.SignOut(uid); err != nil {
		return
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
func (sc *ScoreConfig) GetNormalConfig() NormalCofig {
	c := &NormalCofig{}
	utils.StructCopy(c, sc)
	// log.Debug("check sc:%+v", c)
	return *c
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
		awards = append(awards, item[1])
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
		c := m.Manager.GetNormalConfig()
		if c.MatchID == sc.MatchID {
			isSign = true
		}
	}
	enterTime := ""
	if sc.StartTime > time.Now().Unix() {
		enterTime = time.Unix(sc.StartTime, 0).Format("2006-01-02 15:04:05")
	}

	data := &msg.S2C_RaceDetail{
		ID:            sc.MatchID,
		Desc:          sc.MatchName,
		AwardDesc:     sc.AwardDesc,
		AwardList:     sc.AwardList,
		MatchType:     sc.MatchType,
		RoundNum:      sc.RoundNum,
		EnterTime:     enterTime,
		ConDes:        sc.MatchDesc,
		SignNumDetail: signNumDetail,
		EnterFee:      float64(sc.EnterFee) / 10,
		SignNum:       len(sc.AllSignInPlayers),
		IsSign:        isSign,
	}
	user.WriteMsg(data)
}

// End 赛事結束的逻辑
func (sc *ScoreConfig) End(matchID string) {
	match, ok := MatchList[matchID]
	if !ok {
		return
	}
	for _, p := range match.SignInPlayers {
		sc.RemoveSignPlayer(p)
	}
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
