package match

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"encoding/json"
	"time"

	"github.com/name5566/leaf/timer"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

// 保存所有赛事列表
var (
	MatchList              = map[string]*BaseMatch{}
	UserIDMatch            = map[int]*BaseMatch{}
	UserIDSign             = map[int][]SignList{}
	MatchManagerList       = map[string]values.MatchManager{}
	MatchConfigQueue       = map[string]*values.NormalCofig{} // 后台修改赛事的队列，在比赛开始后下一场生效
	MatchPlayersCountTimer *timer.Timer
)

// SignList 多场次报名列表
type SignList struct {
	MatchID    string
	SonMatchID string `json:"-"`
	MatchName  string
	MatchDesc  string
}

func init() {
	if err := initMatchConfig(); err != nil {
		log.Fatal("init match fail,err:%v", err)
	}
	if err := initMatchTypeConfig(); err != nil {
		log.Fatal("init match config fail,err:%v", err)
	}
	// 启动赛事计时器,一段时间修改假人数
	RefreshPlayersCount()
}

// match
func initMatchConfig() error {
	s := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(s)
	one := map[string]interface{}{}
	log.Debug("init MatchConfig........")
	iter := s.DB(db.DB).C("matchmanager").Find(bson.M{"state": bson.M{"$ne": Delete}}).Iter()
	for iter.Next(&one) {
		if one["matchtype"] == nil || one["matchid"] == nil {
			log.Error("unknow match:%v", one)
			continue
		}
		mType, ok := one["matchtype"].(string)
		if !ok {
			log.Error("unknow match:%v", one)
			continue
		}
		switch mType {
		case ScoreMatch, MoneyMatch, DoubleMatch, QuickMatch:
			sConfig := &ScoreConfig{}
			c, _ := json.Marshal(one)
			if err := json.Unmarshal(c, &sConfig); err != nil {
				log.Error("get config error:%v", err)
				return nil
			}
			if err := sConfig.CheckConfig(); err != nil {
				continue
			}
			if sConfig.UseMatch >= sConfig.TotalMatch {
				game.GetSkeleton().Go(func() {
					db.UpdateMatchManager(sConfig.MatchID, bson.M{"$set": bson.M{"state": End}})
				}, nil)
				continue
			}
			// 上架时间
			if sConfig.ShelfTime > time.Now().Unix() {
				sConfig.State = Cancel
				sConfig.StartTimer = game.GetSkeleton().AfterFunc(time.Duration(sConfig.ShelfTime-time.Now().Unix())*time.Second, func() {
					sConfig.Shelf()
				})
			}
			if sConfig.DownShelfTime > time.Now().Unix() {
				sConfig.DownShelfTimer = game.GetSkeleton().AfterFunc(time.Duration(sConfig.DownShelfTime-time.Now().Unix())*time.Second, func() {
					sConfig.DownShelf()
				})
			}
			sConfig.NewManager()
		default:
			log.Error("unknown match:%v", one)
		}
	}
	log.Debug("init match finish:match manager:%+v", MatchManagerList)
	return nil
}

// initMatchTypeConfig 初始化赛事类型配置
func initMatchTypeConfig() error {
	s := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(s)
	log.Debug("init MatchType Config........")
	one := msg.OneMatchType{}
	// iter := s.DB(db.DB).C("matchtypeconfig").Pipe([]bson.M{
	// 	{"$project": bson.M{
	// 		"_id": 0,
	// 	}},
	// }).Iter()
	iter := s.DB(db.DB).C("matchtypeconfig").Find(bson.M{}).Iter()
	for iter.Next(&one) {
		values.MatchTypeConfig[one.MatchType] = one
	}
	log.Debug("finish init matchtype config %v", values.MatchTypeConfig)

	return nil
}

// GetMatchManagerInfo 获取整个赛事列表的信息
func GetMatchManagerInfo(opt int) interface{} {
	matchManager := []values.MatchManager{}
	for _, v := range MatchManagerList {
		matchManager = append(matchManager, v)
	}
	sortMatch(matchManager)
	switch opt {
	case 1: // 大厅赛事信息推送
		raceInfo := []msg.RaceInfo{}
		for _, v := range matchManager {
			var award float64
			info := v.GetNormalConfig()
			// 不在大厅展示
			if !info.ShowHall || info.State != Signing {
				continue
			}
			if len(info.Award) > 0 {
				for _, v := range info.Award {
					award += values.GetMoneyAward(v)
				}
			}
			sTime := info.StartTime
			if info.StartType == 2 {
				sTime = info.ReadyTime - time.Now().Unix()
			}
			raceInfo = append(raceInfo, msg.RaceInfo{
				ID:           info.MatchID,
				Desc:         info.MatchName,
				Award:        award,
				EnterFee:     float64(info.EnterFee),
				ConDes:       info.MatchDesc,
				JoinNum:      len(info.AllSignInPlayers),
				AllPlayerNum: v.GetAllPlayersCount(),
				StartTime:    sTime,
				StartType:    info.StartType,
				MatchType:    info.MatchType,
				Eliminate:    info.Eliminate,
			})
		}
		return raceInfo
	case 2: // 赛事列表信息推送
		list := []msg.OneMatch{}
		for _, mm := range matchManager {
			m := mm.GetNormalConfig()
			if m.State != Signing {
				continue
			}
			var award float64
			if len(m.Award) > 0 {
				for _, v := range m.Award {
					award += values.GetMoneyAward(v)
				}
			}
			sTime := m.StartTime
			if m.StartType == 2 {
				sTime = m.ReadyTime - time.Now().Unix()
			}
			list = append(list, msg.OneMatch{
				MatchID:      m.MatchID,
				MatchName:    m.MatchName,
				SignInNum:    len(m.AllSignInPlayers),
				AllPlayerNum: mm.GetAllPlayersCount(),
				Recommend:    m.Recommend,
				MaxPlayer:    m.MaxPlayer,
				EnterFee:     m.EnterFee,
				StartTime:    sTime,
				StartType:    m.StartType,
				Award:        award,
				IsSign:       false,
				MatchType:    m.MatchType,
				MatchIcon:    m.MatchIcon,
				ShowHall:     m.ShowHall,
				ConDes:       m.MatchDesc,
				Eliminate:    m.Eliminate,
			})
		}
		return list
	default:
		return nil
	}
}

func sortMatch(list []values.MatchManager) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].GetNormalConfig().Sort > list[j].GetNormalConfig().Sort {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

// BroadcastMatchInfo 当赛事发生变化时，全服广播赛事信息
func BroadcastMatchInfo() {
	raceInfo := GetMatchManagerInfo(2).([]msg.OneMatch)
	// if len(RaceInfo) == 0 {
	// 	return
	// }
	for uid, user := range player.UserIDUsers {
		// myMatchID := ma.NormalCofig.MatchID
		signList := []msg.OneMatch{}
		unsignList := []msg.OneMatch{}
		for _, v := range raceInfo {
			if IsSign(uid, v.MatchID) {
				one := v
				one.IsSign = true
				signList = append(signList, one)
				log.Debug("signList1:%v", signList)
			} else {
				unsignList = append(unsignList, v)
			}
		}

		sendList := append(signList, unsignList...)
		user.WriteMsg(&msg.S2C_GetMatchList{
			List: sendList,
		})
	}
}

// SendMatchInfo 对单个玩家发送
func SendMatchInfo(uid int) {
	raceInfo := GetMatchManagerInfo(2).([]msg.OneMatch)
	user, ok := player.UserIDUsers[uid]
	if !ok {
		return
	}
	signList := []msg.OneMatch{}
	unsignList := []msg.OneMatch{}
	for _, v := range raceInfo {
		if IsSign(uid, v.MatchID) {
			one := v
			one.IsSign = true
			signList = append(signList, one)
			log.Debug("signList1:%v", signList)
		} else {
			unsignList = append(unsignList, v)
		}
	}
	sendList := append(signList, unsignList...)
	user.WriteMsg(&msg.S2C_GetMatchList{
		List: sendList,
	})
}

// IsSign 玩家是否已报名该赛事
func IsSign(uid int, matchID string) bool {
	matchs, ok := UserIDSign[uid]
	if !ok {
		return false
	}
	for _, v := range matchs {
		if v.MatchID == matchID {
			return true
		}
	}
	return false
}

// RefreshPlayersCount 刷新人数
func RefreshPlayersCount() {
	// 首先获取当前时间距离整数点时间差
	_, m, s := time.Now().Clock()
	next := (60-m-1)*60 + (60 - s)
	setFakePlayersCount()
	log.Debug("next:%v", next)
	game.GetSkeleton().AfterFunc(time.Duration(next)*time.Second, RefreshPlayersCount)
}

func setFakePlayersCount() {
	h, _, _ := time.Now().Clock()
	for _, v := range MatchManagerList {
		c := v.GetNormalConfig()
		if !c.ShowHall || c.State != Signing || c.MoneyAward <= 0 {
			v.RefreshFakePlayersCount(0)
			continue
		}
		if c.LimitPlayer == 3 || c.LimitPlayer == 6 || c.LimitPlayer == 9 {
			index := c.LimitPlayer/3 - 1
			v.RefreshFakePlayersCount(FakePlayers[index][h])
		}
	}
}

// GetFakePlayersCount 获取赛事假人数
func GetFakePlayersCount() int {
	count := 0
	for _, v := range MatchManagerList {
		c := v.GetNormalConfig()
		count += c.FakePlayers
	}
	return count
}

// 假人数
var (
	FakePlayers = [][]int{
		[]int{6, 3, 3, 3, 3, 3, 3, 3, 6, 6, 3, 3, 6, 6, 3, 12, 6, 12, 12, 6, 9, 6, 6, 3},
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 6, 6, 13, 18, 13, 13, 17, 12, 6, 19, 17, 12, 12, 18, 12, 6},
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 9, 0, 9, 19, 9, 9, 9, 20, 9, 9, 0},
	}
)
