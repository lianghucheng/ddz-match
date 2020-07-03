package match

import (
	"ddz/game"
	"ddz/game/values"
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
	case Score:
		sConfig := &ScoreConfig{}
		utils.StructCopy(sConfig, data)
		if err := sConfig.CheckConfig(); err != nil {
			code = 1
			desc = "创建赛事参数有误！"
			return
		}
		// 将赛事保存进数据库
		sConfig.Save()
		// 上架时间
		if sConfig.ShelfTime > time.Now().Unix() {
			game.GetSkeleton().AfterFunc(time.Duration(sConfig.ShelfTime-time.Now().Unix())*time.Second, func() {
				NewScoreManager(sConfig)
			})
		} else {
			NewScoreManager(sConfig)
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
	log.Debug("check showhall:%v,%v", c.ShowHall, data.ShowHall)
	if c.ShowHall != data.ShowHall {
		c.ShowHall = data.ShowHall
		m.SetNormalConfig(c)
		m.Save()
		// 通知客户端
		BroadcastMatchInfo()
	}
}

func editMatch(args []interface{}) {
	log.Debug("editMatch:%v", args)
	if len(args) != 1 {
		log.Error("error req:%v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_EditMatch)
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
	if _, ok := MatchManagerList[data.MatchID]; !ok {
		code = 1
		desc = "操作的赛事不存在！"
		return
	}
	c := &values.NormalCofig{}
	utils.StructCopy(c, data)
	MatchConfigQueue[data.MatchID] = c
	// m.SetNormalConfig(c)
	// m.Save()
	// // 通知客户端
	// BroadcastMatchInfo()
}

func optMatch(args []interface{}) {
	log.Debug("optMatch:%+v", args)
	if len(args) != 1 {
		log.Error("error req:%+v", args)
		return
	}
	data, ok := args[0].(*msg.RPC_OptMatch)
	if !ok {
		log.Error("error req:%+v", args)
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
	switch data.Opt {
	case 1: // 上架
		if c.State != Cancel {
			code = 1
			desc = "赛事已上架!"
			return
		}
		c.State = Signing
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
