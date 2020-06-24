package ddz

import (
	. "ddz/game/db"
	. "ddz/game/values"
	"time"

	"github.com/szxby/tools/log"
)

func (game *LandlordMatchRoom) GameDDZRecordInit() {
	for _, userId := range game.PositionUserIDs {
		game.gameRecords[userId] = &DDZGameRecord{
			UserId:    userId,
			MatchId:   game.rule.MatchId,
			Desc:      game.rule.Desc,
			MatchType: game.rule.MatchType,
			Result:    make([]Result, game.rule.Round),
			CreateDat: time.Now().Unix(),
		}
	}
}

func (game *LandlordMatchRoom) GameDDZRecordInsert() {
	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)

	for _, record := range game.gameRecords {
		err := db.DB(DB).C("gamerecord").Insert(record)
		if err != nil {
			log.Error("save gamerecord %v data error: %v", record, err)
		}
	}
}
