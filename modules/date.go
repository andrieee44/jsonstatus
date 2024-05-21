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

func Date(ch chan<- Message, cfg *dateConfig) {
	go sendMessage(ch, "Date", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Date string
			Hour int
		}

		var (
			date     time.Time
			hour     int
			data     jsonStruct
			jsonData json.RawMessage
			err      error
		)

		date = time.Now()
		hour = date.Hour()

		if hour > 12 {
			hour -= 12
		}

		data = jsonStruct{
			Hour: hour,
			Date: date.Format(cfg.Format),
		}

		jsonData, err = json.Marshal(data)
		if err != nil {
			panic(err)
		}

		return jsonData
	})
}
