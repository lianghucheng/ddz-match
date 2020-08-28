package hall

import (
	"ddz/config"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

func init(){
	rand.Seed(time.Now().Unix())
}

func SendHorseRaceLamp(username , matchName ,award string){
	log.Debug("发送跑马灯")
	if len(award) == 0 || values.GetMoneyAward(award) == 0 {
		return
	}

	templates := config.GetCfgNormal().Templates
	if len(templates) == 0{
		log.Error("没有配置跑马灯模版")
		return
	}

	//HorseRaceLamp := &msg.HorseRaceLamp{
	//	UserName:  username,
	//	MatchName: matchName,
	//	Amount:    values.GetMoneyAward(award),
	//}
	//info, err := utils.TransferMapStringSlice(HorseRaceLamp)
	//if err != nil {
	//	log.Error("格式转换失败")
	//	return
	//}
	player.Broadcast(&msg.S2C_HorseRaceLamp{
		Template:  fmt.Sprintf(templates[rand.Intn(len(templates))], username, matchName, values.GetMoneyAward(award)),
		//Info: info,
	})
}