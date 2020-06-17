package utils

import (
	"time"

	"github.com/szxby/tools/log"
)

func HourClock(d time.Duration, fs ...func()) {
	if d < 1*time.Hour {
		log.Error("定时器传入时间不足一小时")
		return
	}
	go func() {
		for {
			for i := 0; i < len(fs); i++ {
				fs[i]()
			}
			now := time.Now()
			next := now.Add(d)
			//下一整点小时
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
			t := time.NewTicker(next.Sub(now))
			<-t.C
		}
	}()
}
