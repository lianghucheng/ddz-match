package game

import (
	. "ddz/game/db"

	"github.com/name5566/leaf/module"
)

var (
	skeleton   = NewSkeleton()
	ChanRPC    = skeleton.ChanRPCServer
	GameModule = new(Module)
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) OnDestroy() {
	MongoDBDestroy()
}

func GetSkeleton() *module.Skeleton {
	return skeleton
}
