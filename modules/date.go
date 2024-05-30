package modules

import (
	"encoding/json"
	"time"
)

type dateConfig struct {
	Enable   bool
	Interval time.Duration
	Format   string
	Icons    []string
}

func date(ch chan<- Message, cfg *dateConfig) {
	go loopMessage(ch, "Date", cfg.Enable, cfg.Interval, func() json.RawMessage {
		var (
			date time.Time
			hour int
		)

		date = time.Now()
		hour = date.Hour()

		if hour > 12 {
			hour -= 12
		}

		return marshalRawJson(struct {
			Icon, Date string
		}{
			Icon: icon(cfg.Icons, 12, float64(hour-1)),
			Date: date.Format(cfg.Format),
		})
	})
}
