package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/szxby/tools/log"
)

type Race struct {
	Races []RaceInfo
}

type RaceInfo struct {
	ID           string
	EnterTime    string
	Desc         string
	EnterFee     float64
	ConDes       string
	Match        int
	MatchType    string
	AwardDesc    string
	AwardTitle   []string
	AwardContent []string
	Award        float64
	RoundNum     string
	Round        int
	BaseScore    int
	LimitPlayer  int
}

var cfg Race

func ReadRaces() {
	cfg = Race{}
	_, err := toml.DecodeFile("conf/ddz.toml", &cfg)
	if err != nil {
		log.Error("读取server.toml失败,error:%v", err)
	}
	log.Release("*****************:%v", cfg)
}

func GetCfgRace() *[]RaceInfo {
	return &cfg.Races
}
