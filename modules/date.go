package modules

import (
	"encoding/json"
	"time"
)

type dateConfig struct {
	Enable   bool
	Interval time.Duration
	Format   string
}

func date(ch chan<- Message, cfg *dateConfig) {
	go loopMessage(ch, "Date", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Date string
			Hour int
		}

		var (
			date time.Time
			hour int
		)

		date = time.Now()
		hour = date.Hour()

		if hour > 12 {
			hour -= 12
		}

		return marshalRawJson(jsonStruct{
			Hour: hour,
			Date: date.Format(cfg.Format),
		})
	})
}
