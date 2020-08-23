package cache

import (
	"ddz/config"
	"ddz/game/db"
	"ddz/game/hall"
	"ddz/game/player"
	"errors"
	"github.com/name5566/leaf/log"
)

func Init() {
	if err := UpdatePropBaseConfig(); err != nil {
		log.Fatal("更新道具基本配置信息失败：" + err.Error())
	}
}

func UpdatePropBaseConfig() error {
	datas := db.ReadPropBaseConfig()
	if len(*datas) == 0 {
		return errors.New("no prop base config. ")
	}
	c := make(map[int]*config.CfgPropBase)
	for _, v := range *datas {
		c[v.PropType] = &config.CfgPropBase{
			PropType: v.PropType,
			Name:     v.Name,
			ImgUrl:   v.ImgUrl,
		}
	}

	config.SetPropBaseConfig(c)

	hall.SendPriceMenu(nil, hall.SendBroacast)
	for _, user := range player.UserIDUsers {
		hall.SendMail(user)
		hall.SendDailySignItems(user)
	}

	return nil
}
