package hall

import (
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"ddz/utils"
	"github.com/szxby/tools/log"
)

func SendActivity(user *player.User, model int) {
	log.Debug("SendActivity")
	datas := db.ReadActivityControl()
	activitys := []msg.ActivityMsg{}
	for _, v := range *datas {
		temp := new(msg.ActivityMsg)
		if err := utils.Transfer(&v, temp); err != nil {
			log.Error(err.Error())
			continue
		}
		activitys = append(activitys, *temp)
	}

	m := &msg.S2C_Activity{
		Datas: activitys,
	}

	if model == 1 {
		log.Debug("single %v", m)
		user.WriteMsg(m)
	} else if model == 2 {
		log.Debug("broadcast %v", m)
		player.Broadcast(m)
	}
}

func SendNotice(user *player.User, model int) {
	log.Debug("SendNotice")
	datas := db.ReadNoticeControl()
	notices := []msg.NoticeMsg{}
	for _, v := range *datas {
		temp := new(msg.NoticeMsg)
		if err := utils.Transfer(&v, temp); err != nil {
			log.Error(err.Error())
			continue
		}
		notices = append(notices, *temp)
	}

	m := &msg.S2C_Notice{
		Datas: notices,
	}

	if model == 1 {
		log.Debug("single %v", m)
		user.WriteMsg(m)
	} else if model == 2 {
		log.Debug("broadcast %v", m)
		player.Broadcast(m)
	}
}

func AddActivityCnt(id int) {
	db.AddCntActivity(id)
}
