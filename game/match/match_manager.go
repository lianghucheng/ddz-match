package match

import (
	"ddz/config"
	"ddz/edy_api"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/hall"
	. "ddz/game/player"
	"ddz/game/values"
	. "ddz/game/values"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// NewScoreManager 创建一个新的比赛类型
// func NewScoreManager(sc *ScoreConfig) {
// 	sc.GetAwardItem()
// 	MatchManagerList[sc.MatchID] = sc
// 	// 如果是倒计时开赛的赛事，一开始直接创建出子比赛
// 	if sc.StartType >= 2 {
// 		sc.CreateOneMatch()
// 	} else {
// 		BroadcastMatchInfo()
// 	}
// }

// NewManager 创建一个新的比赛类型
func (sc *ScoreConfig) NewManager() {
	sc.GetAwardItem()
	MatchManagerList[sc.MatchID] = sc
	if sc.State != Signing {
		return
	}
	// 如果是倒计时开赛的赛事，一开始直接创建出子比赛
	if sc.StartType >= 2 {
		sc.CreateOneMatch()
	} else {
		BroadcastMatchInfo()
	}
}

// SignIn 赛事管理报名
func (sc *ScoreConfig) SignIn(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		log.Error("unknow user:%v", uid)
		return
	}
	if sc.State != Signing {
		user.WriteMsg(&msg.S2C_Apply{
			Error:  msg.S2C_Error_MatchId,
			RaceID: sc.MatchID,
			Action: 1,
			Count:  len(sc.AllSignInPlayers),
		})
		return
	}
	// 当前没有空闲赛事，则创建一个赛事
	if sc.LastMatch == nil || sc.LastMatch.State != Signing {
		sc.CreateOneMatch()
		if sc.LastMatch == nil {
			user.WriteMsg(&msg.S2C_Apply{
				Error:  msg.S2C_Error_MatchId,
				RaceID: sc.MatchID,
				Action: 1,
				Count:  len(sc.AllSignInPlayers),
			})
			return
		}
	}

	// 体总赛事判断分数
	if sc.MatchSource == MatchSourceSportsCenter {
		switch sc.MatchLevel {
		case MatchLevelB:
			if user.BaseData.UserData.SportCenter.RedScore < LevelBSignScore {
				user.WriteMsg(&msg.S2C_Apply{
					Error:  msg.S2C_Error_NotEnoughScore,
					RaceID: sc.MatchID,
					Action: 1,
					Count:  len(sc.AllSignInPlayers),
				})
				return
			}
		case MatchLevelA:
			if user.BaseData.UserData.SportCenter.SilverScore < LevelASignScore {
				user.WriteMsg(&msg.S2C_Apply{
					Error:  msg.S2C_Error_NotEnoughScore,
					RaceID: sc.MatchID,
					Action: 1,
					Count:  len(sc.AllSignInPlayers),
				})
				return
			}
		}
	}
	sc.onSign(uid)
}

func (sc *ScoreConfig) onSign(uid int) {
	user, ok := UserIDUsers[uid]
	if !ok {
		return
	}

	//log.Debug("####机器人报名：%v   %v   %v", user.IsRobot(), sc.RobotNum(), config.GetCfgMatchRobotMaxNums()[sc.MatchID])
	if user.IsRobot() && sc.RobotNum() > config.GetCfgMatchRobotMaxNums()[sc.MatchID]-1 {
		user.WriteMsg(&msg.S2C_Apply{
			Error:  msg.S2C_Error_MoreRobot,
			RaceID: sc.MatchID,
			Action: 1,
			Count:  len(sc.AllSignInPlayers),
		})
		return
	}

	// ok 先将报名玩家添加进去,比赛开始后将触发清除报名玩家的逻辑
	sc.AllSignInPlayers = append(sc.AllSignInPlayers, uid)
	if err := sc.LastMatch.SignIn(uid); err != nil {
		sc.AllSignInPlayers = sc.AllSignInPlayers[:len(sc.AllSignInPlayers)-1]
		return
	}

	// 报名满人则清零
	// if len(sc.AllSignInPlayers) >= sc.MaxPlayer && sc.StartType == 1 {
	// 	sc.AllSignInPlayers = []int{}
	// }

	sc.AllPlayingPlayersCount++

	user.WriteMsg(&msg.S2C_Apply{
		Error:  0,
		RaceID: sc.MatchID,
		Action: 1,
		Count:  len(sc.AllSignInPlayers),
	})

	Broadcast(&msg.S2C_MatchNum{
		MatchId:      sc.MatchID,
		Count:        len(sc.AllSignInPlayers),
		AllPlayerNum: sc.GetAllPlayersCount(),
	})

	// 刷新玩家赛事列表
	SendMatchInfo(uid)

	// 赛事结束
	if sc.UseMatch >= sc.TotalMatch && sc.LastMatch.State != Signing {
		sc.EndMatchManager()
	}
}

// SignOut 赛事管理报名
func (sc *ScoreConfig) SignOut(uid int, matchID string) {
	user, ok := UserIDUsers[uid]
	// if !ok {
	// 	log.Error("unknow user:%v", uid)
	// 	return
	// }

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
		hall.MatchInterruptPushMail(uid, sc.MatchName, sc.EnterFee)
	}
	sc.AllPlayingPlayersCount--
	// ok
	sc.RemoveSignPlayer(uid)
	if ok {
		user.WriteMsg(&msg.S2C_Apply{
			Error:  0,
			RaceID: sc.MatchID,
			Action: 2,
			Count:  len(sc.AllSignInPlayers),
		})
	}
	Broadcast(&msg.S2C_MatchNum{
		MatchId:      sc.MatchID,
		Count:        len(sc.AllSignInPlayers),
		AllPlayerNum: sc.GetAllPlayersCount(),
	})

	// 如果当前赛事玩家全部退出，那么检查一次是否有新赛事配置更新
	if len(sc.AllSignInPlayers) == 0 {
		sc.CheckNewConfig()
	}
}

// GetNormalConfig 获取通用配置
func (sc *ScoreConfig) GetNormalConfig() *NormalCofig {
	c := &NormalCofig{}
	utils.StructCopy(c, sc)
	// log.Debug("check sc:%+v", c)
	return c
}

// SetNormalConfig 获取通用配置
func (sc *ScoreConfig) SetNormalConfig(config *NormalCofig) {
	// 部分参数不配置
	config.AllSignInPlayers = sc.AllSignInPlayers
	config.AllPlayingPlayersCount = sc.AllPlayingPlayersCount
	utils.StructCopy(sc, config)
	sc.CheckConfig()
	if len(config.AwardList) > 0 {
		// 重新解析award
		sc.GetAwardItem()
	}
}

// CheckNewConfig 检查是否有新的配置需要更新
func (sc *ScoreConfig) CheckNewConfig() {
	if c, ok := MatchConfigQueue[sc.MatchID]; ok {
		sc.SetNormalConfig(c)
		sc.Save()
		// 通知客户端
		BroadcastMatchInfo()
		delete(MatchConfigQueue, sc.MatchID)
	}
}

// GetAwardItem 根据list，解析出具体的奖励物品
func (sc *ScoreConfig) GetAwardItem() {
	list := sc.AwardList
	tmp := map[string]interface{}{}
	err := json.Unmarshal([]byte(list), &tmp)
	if err != nil {
		log.Error("unmarshal fail %v", err)
		return
	}
	awards := make([]string, len(tmp))
	sc.MoneyAward = 0
	sc.CouponAward = 0
	sc.FragmentAward = 0
	for i, v := range tmp {
		num := []byte{}
		for _, s := range []byte(i) {
			if s <= 57 && s >= 46 {
				num = append(num, s)
			}
		}
		rank, err := strconv.Atoi(string(num))
		if err != nil {
			continue
		}
		// log.Debug("check:%v", rank)
		award, ok := v.(string)
		if !ok {
			continue
		}
		if rank-1 < 0 || rank > len(awards) {
			continue
		}
		// log.Debug("award:%v", award)
		awards[rank-1] = award
		sc.MoneyAward += values.GetMoneyAward(award)
		sc.CouponAward += values.GetCouponAward(award)
		sc.FragmentAward += values.GetFragmentAward(award)
	}
	sc.Award = awards
	log.Debug("match award:%v", sc.Award)
	// log.Debug("items:%v", getNoneScoreAward(sc.Award))
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
	// if m, ok := UserIDMatch[uid]; ok {
	// 	if m.NormalCofig.MatchID == sc.MatchID {
	// 		isSign = true
	// 	}
	// }
	if matchs, ok := UserIDSign[uid]; ok {
		for _, v := range matchs {
			if v.MatchID == sc.MatchID {
				isSign = true
				break
			}
		}
	}
	sTime := sc.StartTime
	if sc.StartType == 2 {
		sTime = sc.ReadyTime - time.Now().Unix()
	}
	data := &msg.S2C_RaceDetail{
		ID:        sc.MatchID,
		Desc:      sc.MatchName,
		AwardDesc: sc.AwardDesc,
		// AwardList:     sc.AwardList,
		AwardList:     getNoneScoreAward(sc.Award),
		MatchType:     sc.MatchType,
		RoundNum:      sc.RoundNum,
		StartTime:     sTime,
		StartType:     sc.StartType,
		ConDes:        sc.MatchDesc,
		SignNumDetail: signNumDetail,
		EnterFee:      float64(sc.EnterFee),
		SignNum:       len(sc.AllSignInPlayers),
		IsSign:        isSign,
		Eliminate:     sc.Eliminate,
	}
	user.WriteMsg(data)
}

// End 赛事結束的逻辑
func (sc *ScoreConfig) End(matchID string) {
	match, ok := MatchList[matchID]
	if !ok {
		return
	}
	log.Debug("all:%v,sign:%+v", sc.AllPlayingPlayersCount, match.SignInPlayers)
	sc.AllPlayingPlayersCount -= len(match.SignInPlayers)
	// 广播比赛剩余人数
	Broadcast(&msg.S2C_MatchNum{
		MatchId:      sc.MatchID,
		Count:        len(sc.AllSignInPlayers),
		AllPlayerNum: sc.GetAllPlayersCount(),
	})

	// 如果是后台下架赛事,赛事使用次数返回
	if match.OptMatchType == 2 {
		match.OptMatchType = 0
		sc.UseMatch--
	}

	// for _, p := range match.SignInPlayers {
	// 	sc.RemoveSignPlayer(p)
	// }
	game.GetSkeleton().Go(func() {
		db.UpdateMatchManager(sc.MatchID, bson.M{"$set": bson.M{"usematch": sc.UseMatch}})
	}, nil)
}

// RemoveSignPlayer 清除签到玩家
func (sc *ScoreConfig) RemoveSignPlayer(uid int) {
	log.Debug("before remove uid:%v,%+v", uid, sc.AllSignInPlayers)
	for i, p := range sc.AllSignInPlayers {
		if p != uid {
			continue
		}
		if i == len(sc.AllSignInPlayers)-1 {
			sc.AllSignInPlayers = sc.AllSignInPlayers[:i]
		} else {
			sc.AllSignInPlayers = append(sc.AllSignInPlayers[:i], sc.AllSignInPlayers[i+1:]...)
		}
		break
	}
	log.Debug("after remove uid:%v,%+v", uid, sc.AllSignInPlayers)
}

// CreateOneMatch 创建子赛事
func (sc *ScoreConfig) CreateOneMatch() {
	if sc.UseMatch >= sc.TotalMatch {
		sc.EndMatchManager()
		return
	}
	sonID := ""
	// if sc.CurrentIDIndex < len(sc.OfficalIDs) {
	if sc.MatchSource == MatchSourceSportsCenter {
		sonID = sc.MatchID + "01" + fmt.Sprintf("%06d", (sc.UseMatch+1))
	} else {
		sonID = sc.MatchID + strconv.FormatInt(time.Now().Unix(), 10)
	}
	log.Debug("usematch:%v,sonmatchid:%v", sc.UseMatch, sonID)
	sc.SonMatchID = sonID
	if sc.StartType == 2 {
		sc.ReadyTime = time.Now().Unix() + sc.StartTime
	} else if sc.StartType == 3 {
		if sc.UseMatch >= sc.TotalMatch {
			return
		}
		// 第三种赛事为每天固定时间开一次赛
		if sc.StartTime < time.Now().Unix()+1 {
			var oneDay int64 = 24 * 60 * 60
			for start := sc.StartTime; start > 0; start += oneDay {
				if start > time.Now().Unix() {
					sc.StartTime = start
					break
				}
			}
		}
		sc.Save()
	}
	// nSconfig := &ScoreConfig{}
	// utils.StructCopy(nSconfig, sc)
	newMatch := NewScoreMatch(sc)
	// newMatch.Manager = sc
	sc.LastMatch = newMatch
	sc.UseMatch++
	sc.CurrentIDIndex++

	BroadcastMatchInfo()

	if sc.MatchSource == MatchSourceSportsCenter {
		var createErr error
		game.GetSkeleton().Go(func() {
			_, createErr = edy_api.CheckMatch(sc.LastMatch.SonMatchID)
		}, func() {
			if createErr != nil {
				sc.DownShelf()
			}
		})
	}
}

// Save 保存赛事
func (sc *ScoreConfig) Save() error {
	log.Debug("check config:%+v", sc)
	err := db.UpdateMatchManager(sc.MatchID, sc)
	return err
}

// EndMatchManager 该类赛事结束
func (sc *ScoreConfig) EndMatchManager() {
	delete(MatchManagerList, sc.MatchID)
	// 通知客户端
	BroadcastMatchInfo()
	// 修改赛事配置数据
	game.GetSkeleton().Go(func() {
		db.UpdateMatchManager(sc.MatchID, bson.M{"$set": bson.M{"state": Delete}})
	}, nil)
}

// CheckConfig 检查后台传输的配置文件，查漏补缺
func (sc *ScoreConfig) CheckConfig() error {
	if sc.StartType == 0 || sc.Round == 0 || len(sc.AwardList) == 0 || len(sc.MatchID) == 0 || sc.LimitPlayer == 0 ||
		len(sc.Recommend) == 0 || sc.TotalMatch == 0 || len(sc.MatchName) == 0 || len(sc.MatchType) == 0 ||
		// (sc.StartType == 3 && sc.StartTime < time.Now().Unix()) ||
		(sc.StartType == 2 && sc.StartTime <= 0) || sc.TotalMatch < sc.UseMatch {
		log.Error("invalid config:%+v", sc)
		return errors.New("config error")
	}
	if sc.MaxPlayer == 0 && sc.StartType == 1 {
		sc.MaxPlayer = sc.LimitPlayer
	}
	if sc.TablePlayer == 0 {
		sc.TablePlayer = 3
	}
	if len(sc.AwardDesc) == 0 {
		sc.GetAwardItem()
		sc.AwardDesc = fmt.Sprintf("前%v名有奖励", len(sc.Award))
	}
	if sc.StartType == 1 {
		sc.MaxPlayer = sc.LimitPlayer
	}
	if sc.Card == 0 {
		sc.Card = 1
	}
	// if len(sc.MatchDesc) == 0 {
	if sc.StartType == 1 {
		sc.MatchDesc = fmt.Sprintf("满%v人开赛", sc.LimitPlayer)
	} else {
		sc.MatchDesc = fmt.Sprintf("%v人", sc.LimitPlayer)
	}
	// }
	if sc.BaseScore == 0 {
		sc.BaseScore = 1
	}

	// 刷新描述
	sc.RoundNum = fmt.Sprintf("%v轮%v副", sc.Round, sc.Card)
	return nil
}

// ClearLastMatch 清理最近一场比赛
func (sc *ScoreConfig) ClearLastMatch() {
	sc.LastMatch = nil
	sc.SonMatchID = ""
}

func (sc *ScoreConfig) RobotNum() int {
	cnt := 0
	for _, v := range sc.AllSignInPlayers {
		ud := ReadUserDataByID(v)
		if ud.Role == RoleRobot {
			cnt++
		}
	}
	return cnt
}

// CloseMatch 系统停服维护,踢出所有报名玩家
func (sc *ScoreConfig) CloseMatch() {
	log.Debug("last:%+v", sc.LastMatch)
	if sc.LastMatch != nil && sc.LastMatch.State == Signing {
		for _, p := range sc.LastMatch.AllPlayers {
			sc.SignOut(p.BaseData.UserData.UserID, sc.MatchID)
			hall.GamePushMail(p.BaseData.UserData.UserID, "更新通知",
				fmt.Sprintf("平台于[%v]－[%v]进行优化更新,期间暂时无法登录游戏,请您谅解。祝各位选手玩的开心，赢得漂亮！",
					time.Unix(values.DefaultRestartConfig.RestartTime, 0).Format("2006-01-02 15:04:05"),
					time.Unix(values.DefaultRestartConfig.EndTime, 0).Format("2006-01-02 15:04:05")))
		}
	}
}

// SetStartTimer 设置开始时间
func (sc *ScoreConfig) SetStartTimer(timer *timer.Timer) {
	if sc.StartTimer != nil {
		sc.StartTimer.Stop()
	}
	sc.StartTimer = timer
}

// SetDownShelfTimer 设置开始时间
func (sc *ScoreConfig) SetDownShelfTimer(timer *timer.Timer) {
	if sc.DownShelfTimer != nil {
		sc.DownShelfTimer.Stop()
	}
	sc.DownShelfTimer = timer
}

// stopAllTimer 停止所有计时器
func (sc *ScoreConfig) stopAllTimer() {
	log.Debug("stop timer:%v", sc.StartTimer)
	if sc.StartTimer != nil {
		sc.StartTimer.Stop()
	}
	if sc.DownShelfTimer != nil {
		sc.DownShelfTimer.Stop()
	}
}

// DownShelf 下架
func (sc *ScoreConfig) DownShelf() {
	if sc.State != Signing {
		log.Error("down shelf unCancel matchID:%v", sc.MatchID)
		return
	}

	if sc.SonMatchID != "" {
		match, ok := MatchList[sc.SonMatchID]
		if ok && match.State == Signing {
			match.OptMatchType = 2
			match.CloseMatch()
		}
	}
	sc.stopAllTimer()
	sc.State = Cancel

	// 刷新数据库
	sc.Save()
	if sc.ShowHall {
		// 通知客户端
		BroadcastMatchInfo()
	}
}

// Shelf 上架
func (sc *ScoreConfig) Shelf() {
	if sc.State != Cancel {
		log.Error("shelf not cancel matchID:%v", sc.MatchID)
		return
	}
	sc.State = Signing

	if sc.StartTimer != nil {
		sc.StartTimer.Stop()
		switch sc.MatchType {
		case ScoreMatch, MoneyMatch, DoubleMatch, QuickMatch:
			sc.NewManager()
		default:
			log.Error("unknown match:%+v", sc)
		}
	}

	// 刷新数据库
	sc.Save()
	if sc.ShowHall {
		// 通知客户端
		BroadcastMatchInfo()
	}
}

// Delete 删除
func (sc *ScoreConfig) Delete() {
	if sc.State != Cancel {
		log.Error("delete not cancel matchID:%v", sc.MatchID)
		return
	}

	sc.stopAllTimer()
	sc.State = Delete
	delete(MatchManagerList, sc.MatchID)

	// 刷新数据库
	sc.Save()
}

// GetAllPlayersCount 获取赛事总人数(包括假人数)
func (sc *ScoreConfig) GetAllPlayersCount() int {
	return sc.AllPlayingPlayersCount + sc.FakePlayers
}

// RefreshFakePlayersCount 刷新赛事总人数(包括假人数)
func (sc *ScoreConfig) RefreshFakePlayersCount(count int) {
	log.Debug("match %v refresh fake %v", sc.MatchName, count)
	sc.FakePlayers = count
}
