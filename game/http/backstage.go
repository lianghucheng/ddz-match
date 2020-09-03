package http

import (
	"crypto/sha256"
	"ddz/game"
	"ddz/msg"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/szxby/tools/log"
)

const (
	key = "7inrmpd5DSQTfDxnAnOH"
)

// CalculateHash calculate hash
func CalculateHash(data string) string {
	h := sha256.New()
	h.Write([]byte(key + data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

type rawPack struct {
	Sign string
	Data string
}

func checkSignature(msg []byte) bool {
	pkg := rawPack{}
	if err := json.Unmarshal(msg, &pkg); err != nil {
		log.Error("umarshal msg fail %v", err)
		return false
	}
	sign := pkg.Sign
	data := pkg.Data
	// log.Debug("signData:%v", data)
	// log.Debug("sign:%v", signature(data))
	if CalculateHash(data) != sign {
		return false
	}
	return true
}

// 统一的解包方法
func unpack(rawBody io.ReadCloser) string {
	body, err := ioutil.ReadAll(rawBody)
	defer rawBody.Close()
	if err != nil {
		log.Error("unpack fail:%v", err)
		return ""
	}
	log.Debug("unpack data:%v", string(body))

	pkg := rawPack{}
	if err := json.Unmarshal(body, &pkg); err != nil {
		log.Error("umarshal msg fail %v", err)
		return ""
	}
	sign := pkg.Sign
	data := pkg.Data

	if CalculateHash(data) != sign {
		log.Error("check sign fail")
		return ""
	}
	return data
}

func addMatch(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	add := msg.RPC_AddManagerReq{}
	if err := json.Unmarshal([]byte(ret), &add); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	add.WG = &wg
	add.Write = w
	game.ChanRPC.Go("addMatch", &add)
	wg.Wait()
}

func showHall(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	show := msg.RPC_ShowHall{}
	if err := json.Unmarshal([]byte(ret), &show); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	show.WG = &wg
	show.Write = w
	game.ChanRPC.Go("showHall", &show)
	wg.Wait()
}

func editSort(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	data := msg.RPC_EditSort{}
	if err := json.Unmarshal([]byte(ret), &data); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	data.WG = &wg
	data.Write = w
	game.ChanRPC.Go("editSort", &data)
	wg.Wait()
}

func editMatch(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	edit := msg.RPC_EditMatch{}
	if err := json.Unmarshal([]byte(ret), &edit); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	log.Debug("check:%+v", edit)
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	edit.WG = &wg
	edit.Write = w
	game.ChanRPC.Go("editMatch", &edit)
	wg.Wait()
}

func optMatch(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	opt := msg.RPC_OptMatch{}
	if err := json.Unmarshal([]byte(ret), &opt); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	opt.WG = &wg
	opt.Write = w
	game.ChanRPC.Go("optMatch", &opt)
	wg.Wait()
}

func optUser(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	opt := msg.RPC_OptUser{}
	if err := json.Unmarshal([]byte(ret), &opt); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	opt.WG = &wg
	opt.Write = w
	game.ChanRPC.Go("optUser", &opt)
	wg.Wait()
}

func clearRealInfo(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	opt := msg.RPC_ClearInfo{}
	if err := json.Unmarshal([]byte(ret), &opt); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	opt.WG = &wg
	opt.Write = w
	game.ChanRPC.Go("clearInfo", &opt)
	wg.Wait()
}

func restartServer(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	data := msg.RPC_Restart{}
	if err := json.Unmarshal([]byte(ret), &data); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	data.WG = &wg
	data.Write = w
	game.ChanRPC.Go("restartServer", &data)
	wg.Wait()
}

func editWhiteList(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	data := msg.RPC_EditWhiteList{}
	// if err := json.Unmarshal([]byte(ret), &data); err != nil {
	// 	code = 1
	// 	desc = "请求参数有误！"
	// 	resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
	// 	w.Write(resp)
	// 	return
	// }
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	data.WG = &wg
	data.Write = w
	game.ChanRPC.Go("editWhiteList", &data)
	wg.Wait()
}

// 获取在线人数
func getOnline(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	data := msg.RPC_GetOnline{}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	data.WG = &wg
	data.Write = w
	game.ChanRPC.Go("getOnline", &data)
	wg.Wait()
}

// 重新回调异常赛事
func dealIllegalMatch(w http.ResponseWriter, req *http.Request) {
	ret := unpack(req.Body)
	code := 0
	desc := "OK"
	if ret == "" {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	data := msg.RPC_IllegalMatch{}
	if err := json.Unmarshal([]byte(ret), &data); err != nil {
		code = 1
		desc = "请求参数有误！"
		resp, _ := json.Marshal(map[string]interface{}{"code": code, "desc": desc})
		w.Write(resp)
		return
	}
	// 等待主协程处理完成后返回
	wg := sync.WaitGroup{}
	wg.Add(1)
	data.WG = &wg
	data.Write = w
	game.ChanRPC.Go("dealIllegalMatch", &data)
	wg.Wait()
}
