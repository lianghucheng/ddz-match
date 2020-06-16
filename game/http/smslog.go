package http

import (
	. "ddz/game/db"
)

var smsJuHeLogChan chan *JuHeSmsLog

func init() {

	go InsertJuHeSmsLog()
}

func writeJuHeSmsLog(data *JuHeSmsLog) {
	if smsJuHeLogChan == nil {
		return
	}
	if len(smsJuHeLogChan) >= 10000 {
		return
	}
	smsJuHeLogChan <- data
}

func InsertJuHeSmsLog() {
	smsJuHeLogChan = make(chan *JuHeSmsLog, 10000)
	for smslog := range smsJuHeLogChan {
		if smslog == nil {
			break
		}
		err := insertJuHe(smslog)
		if err != nil {
			for i := 0; i < 5; i++ {
				err := insertJuHe(smslog)
				if err != nil {
					break
				}
			}
		}
	}
	close(smsJuHeLogChan)
}

func insertJuHe(smslog *JuHeSmsLog) error {
	se := MongoDB.Ref()
	defer MongoDB.UnRef(se)
	err := se.DB(DB).C("juhesmslog").Insert(smslog)
	return err
}
