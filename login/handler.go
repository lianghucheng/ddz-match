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
	handler(&msg.C2S_UsrnPwdLogin{}, handleUsrnPwdLogin)
	handler(&msg.C2S_Register{}, handleRegister)
	handler(&msg.C2S_FindPassword{}, handleFindPassword)
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
	//login
	game.ChanRPC.Go("AccountLogin", a, m)
}

func handleUsrnPwdLogin(args []interface{}) {
	m := args[0].(*msg.C2S_UsrnPwdLogin)
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("UsernamePasswordLogin", a, m)
}

func handleRegister(args []interface{}) {
	m := args[0].(*msg.C2S_Register)
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("Register", a, m)
}

func handleFindPassword(args []interface{}) {
	m := args[0].(*msg.C2S_FindPassword)
	a := args[1].(gate.Agent)

	game.ChanRPC.Go("FindPassword", a, m)
}
