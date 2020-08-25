package rpc

import (
	"container/list"
	"ddz/conf"
	"ddz/game"
	"errors"
	"net/rpc"
	"time"

	"github.com/name5566/leaf/timer"

	"github.com/szxby/tools/log"
)

var (
	activityClient *rpc.Client
	dailyDataQueue = struct {
		list *list.List
		RWC  chan bool
	}{
		list: list.New(),
		RWC:  make(chan bool, 1),
	}
	connectTimer *timer.Timer
	// pushQuene=make
)

func init() {
	connectToActivityServer()
}

func connectToActivityServer() {
	dailyDataQueue.RWC <- true
	defer func() {
		<-dailyDataQueue.RWC
	}()
	log.Debug("connect to activity server......")
	if connectTimer != nil {
		connectTimer.Stop()
	}
	client, err := rpc.DialHTTP("tcp", conf.GetCfgLeafSrv().ActivityServer)
	if err != nil {
		log.Debug("dialing:%v", err)
		connectTimer = game.GetSkeleton().AfterFunc(5*time.Second, connectToActivityServer)
		return
	}
	activityClient = client
	// log.Debug("call data:%v", dailyDataQueue.list.Front())
	for dailyDataQueue.list.Front() != nil {
		e := dailyDataQueue.list.Front()
		req := e.Value.(*RPCReq)
		if err := CallActivityServer(req.Method, req.Send, nil); err != nil {
			log.Error("err:%v", err)
		}
		dailyDataQueue.list.Remove(e)
	}
	connectTimer = nil
}

// PushData 当请求失败时,存在队列中,请求成功后再一次发送
func PushData(data interface{}) {
	log.Debug("pushData:%v", data)
	game.GetSkeleton().Go(func() {
		dailyDataQueue.RWC <- true
		defer func() {
			<-dailyDataQueue.RWC
		}()
		dailyDataQueue.list.PushBack(data)
		log.Debug("data:%v", dailyDataQueue.list.Front())
	}, nil)
}

// CallActivityServer 向活动服发送数据
func CallActivityServer(method string, send interface{}, reply *RPCRet) error {
	log.Debug("call activity:%v,%v,%v", method, send, reply)
	req := &RPCReq{
		Method: method,
		Send:   send,
	}
	push := false
	if reply == nil {
		push = true
		reply = &RPCRet{}
	}
	if activityClient == nil {
		if connectTimer == nil {
			connectToActivityServer()
		}
		if push {
			PushData(req)
		}
		return errors.New("call client err")
	}
	err := activityClient.Call(method, send, reply)
	if err == rpc.ErrShutdown {
		if push {
			PushData(req)
		}
		if connectTimer == nil {
			connectToActivityServer()
		}
		log.Error("err:%v", err)
		return err
	}
	if reply.Code != 0 {
		return errors.New(reply.Desc)
	}
	return nil
}

// RPCRet 统一返回
type RPCRet struct {
	Code int
	Desc string
	Data interface{}
}

// RPCReq 统一请求对象
type RPCReq struct {
	Method string
	Send   interface{}
}

// RPCUploadMatchInfo 上传比赛结果
type RPCUploadMatchInfo struct {
	AccountID int
	OptTime   int64
}

// RPCGetDailyWelfareInfo 获取每日福利详情
type RPCGetDailyWelfareInfo struct {
	AccountID int
}

// RPCDrawDailyWelfare 玩家领取每日福利
type RPCDrawDailyWelfare struct {
	AccountID  int
	DailyType  int // 奖励类型
	AwardIndex int // 领取奖励序列号
}
