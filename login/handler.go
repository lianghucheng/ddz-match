package login

import (
	"ddz/game"
	"ddz/msg"
	"reflect"

	"github.com/name5566/leaf/gate"
)

func init() {
	handler(&msg.C2S_TokenLogin{}, handleC2STokenLogin)
	handler(&msg.C2S_AccountLogin{}, handleC2SUsernamePasswordLogin)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)

}

func handleC2STokenLogin(args []interface{}) {
	// 收到的 C2S_TokenLogin 消息
	m := args[0].(*msg.C2S_TokenLogin)
	// 消息的发送者
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("TokenLogin", a, m)
}

func handleC2SUsernamePasswordLogin(args []interface{}) {
	m := args[0].(*msg.C2S_AccountLogin)
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("UsernamePasswordLogin", a, m)
}
