package login

import (
	"github.com/name5566/leaf/module"
)

var (
	skeleton    = NewSkeleton()
	ChanRPC     = skeleton.ChanRPCServer
	LoginModule = new(Module)
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) OnDestroy() {

}
