package main

import (
	"ddz/conf"
	"ddz/game"
	_ "ddz/game/http"
	_ "ddz/game/match"
	_ "ddz/game/session"
	"ddz/gate"
	"ddz/login"
	"log"

	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	//conf2.DBCfgInit()
	lconf.LogLevel = conf.GetCfgLeafSrv().LogLevel
	lconf.LogPath = conf.GetCfgLeafSrv().LogPath
	lconf.LogFlag = log.Lshortfile | log.LstdFlags
	lconf.ConsolePort = conf.GetCfgLeafSrv().ConsolePort
	lconf.ProfilePath = conf.GetCfgLeafSrv().ProfilePath
	leaf.Run(
		game.GameModule,
		gate.ModuleGate,
		login.LoginModule,
	)
}
