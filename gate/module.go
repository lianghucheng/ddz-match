package gate

import (
	"ddz/conf"
	"ddz/game"
	"ddz/msg"
	"time"

	"github.com/name5566/leaf/gate"
)

type Module struct {
	*gate.Gate
}

var ModuleGate = new(Module)

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.GetCfgLeafSrv().MaxConnNum,
		PendingWriteNum: 2000,
		MaxMsgLen:       8192,
		WSAddr:          conf.GetCfgLeafSrv().WSAddr,
		HTTPTimeout:     10 * time.Second,
		CertFile:        conf.GetCfgLeafSrv().CertFile,
		KeyFile:         conf.GetCfgLeafSrv().KeyFile,
		TCPAddr:         conf.GetCfgLeafSrv().TCPAddr,
		LenMsgLen:       2,
		LittleEndian:    false,
		Processor:       msg.Processor,
		AgentChanRPC:    game.ChanRPC,
	}
}
