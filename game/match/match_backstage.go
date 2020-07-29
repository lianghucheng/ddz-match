package match

import (
	"ddz/game"
	"ddz/msg"
	"ddz/utils"
	"encoding/json"
	"time"

	"github.com/szxby/tools/log"
)

func init() {
	game.GetSkeleton().RegisterChanRPC("addMatch", addMatch)   // 新增赛事
	game.GetSkeleton().RegisterChanRPC("showHall", showHall)   // 控制某一赛事是否在大厅显示
	game.GetSkeleton().RegisterChanRPC("editMatch", editMatch) // 配置赛事
	game.GetSkeleton().RegisterChanRPC("optMatch", optMatch)   // 操作赛事，1上架，2下架，3删除
	game.GetSkeleton().RegisterChanRPC("editSort", editSort)   // 修改赛事排序
}

func addMatch(args []interface{}) {
	log.Debug("addMatch:%v", args)
	if len(args) != 1 {
		log.Error("error req:%v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_AddManagerReq)
	if !ok {
		log.Error("error req:%v", args)
		return
	}
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	switch data.MatchType {
	case ScoreMatch, MoneyMatch, DoubleMatch, QuickMatch:
		sConfig := &ScoreConfig{}
		utils.StructCopy(sConfig, data)
		if err := sConfig.CheckConfig(); err != nil {
			code = 1
			desc = "创建赛事参数有误！"
			return
		}
		if _, ok := MatchManagerList[sConfig.MatchID]; ok {
			code = 1
			desc = "赛事ID重复！"
			return
		}
		if sConfig.ShelfTime > time.Now().Unix() {
			sConfig.State = Cancel
		}
		// 将赛事保存进数据库
		if err := sConfig.Save(); err != nil {
			code = 1
			desc = "创建赛事失败！"
			return
		}
		// 上架时间
		if sConfig.ShelfTime > time.Now().Unix() {
			sConfig.StartTimer = game.GetSkeleton().AfterFunc(time.Duration(sConfig.ShelfTime-time.Now().Unix())*time.Second, func() {
				// NewScoreManager(sConfig)
				sConfig.NewManager()
			})
			MatchManagerList[sConfig.MatchID] = sConfig
		} else {
			// NewScoreManager(sConfig)
			sConfig.NewManager()
		}
	default:
		code = 1
		desc = "新增赛事未知，请重新确认！"
		log.Error("unknown match:%v", data)
		return
	}
}

func showHall(args []interface{}) {
	log.Debug("showhall:%v", args)
	if len(args) != 1 {
		log.Error("error req:%v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_ShowHall)
	if !ok {
		log.Error("error req:%v", args)
		return
	}
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	m, ok := MatchManagerList[data.MatchID]
	if !ok {
		code = 1
		desc = "操作的赛事不存在！"
		return
	}
	c := m.GetNormalConfig()
	if c.ShowHall != data.ShowHall {
		c.ShowHall = data.ShowHall
		m.SetNormalConfig(c)
		m.Save()
		// 通知客户端
		BroadcastMatchInfo()
	}
}

func editSort(args []interface{}) {
	log.Debug("showhall:%v", args)
	if len(args) != 1 {
		log.Error("error req:%v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_EditSort)
	if !ok {
		log.Error("error req:%v", args)
		return
	}
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	m, ok := MatchManagerList[data.MatchID]
	if !ok {
		code = 1
		desc = "操作的赛事不存在！"
		return
	}
	c := m.GetNormalConfig()
	if c.Sort != data.Sort {
		c.Sort = data.Sort
		m.SetNormalConfig(c)
		m.Save()
		// 通知客户端
		BroadcastMatchInfo()
	}
}

func editMatch(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_EditMatch)
	if !ok {
		log.Error("error req:%v", args)
		return
	}
	log.Debug("editMatch:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	m, ok := MatchManagerList[data.MatchID]
	if !ok {
		code = 1
		desc = "操作的赛事不存在！"
		return
	}
	c := m.GetNormalConfig()
	utils.StructCopy(c, data)

	if c.State > Cancel {
		code = 1
		desc = "当前赛事不可修改!"
		return
	}

	// 当前赛事没人，且处于正常状态则直接修改
	if len(c.AllSignInPlayers) == 0 {
		m.SetNormalConfig(c)
		m.Save()
		// 通知客户端
		BroadcastMatchInfo()
	} else {
		c.AllSignInPlayers = []int{}
		MatchConfigQueue[data.MatchID] = c
	}
}

func optMatch(args []interface{}) {
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_OptMatch)
	if !ok {
		log.Error("error req:%+v", args)
		return
	}
	log.Debug("optMatch:%+v", data)
	code := 0
	desc := "OK"
	defer func() {
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		data.Write.Write(resp)
		data.WG.Done()
	}()
	m, ok := MatchManagerList[data.MatchID]
	if !ok {
		code = 1
		desc = "操作的赛事不存在！"
		return
	}

	c := m.GetNormalConfig()
	switch data.Opt {
	case 1: // 上架
		if c.State != Cancel {
			code = 1
			desc = "赛事已上架!"
			return
		}
		c.State = Signing
		if c.StartTimer != nil {
			c.StartTimer.Stop()
			switch c.MatchType {
			case ScoreMatch, MoneyMatch, DoubleMatch, QuickMatch:
				m.NewManager()
			default:
				log.Error("unknown match:%+v", c)
			}
		}
	case 2: // 下架
		if c.State != Signing {
			code = 1
			desc = "赛事未上架!"
			return
		}
		c.State = Cancel
		if c.SonMatchID != "" {
			match, ok := MatchList[c.SonMatchID]
			if ok && match.State == Signing {
				match.CloseMatch()
			}
		}
	case 3: // 删除
		if c.State < Cancel {
			code = 1
			desc = "赛事未下架!"
			return
		}
		if c.StartTimer != nil {
			c.StartTimer.Stop()
		}
		c.State = Delete
		delete(MatchManagerList, c.MatchID)
	default: // 未知
		log.Error("unknown opt:%v", data)
		code = 1
		desc = "未知操作！"
		return
	}
	m.SetNormalConfig(c)
	// 刷新数据库
	m.Save()
	// 通知客户端
	BroadcastMatchInfo()
}
