package modules

import (
	"encoding/json"
	"time"

	"github.com/BurntSushi/toml"
)

func Date(ch chan<- Message, decoder *toml.Decoder) {
	type config struct {
		Enable   bool          `toml:"date_enable"`
		Interval time.Duration `toml:"date_interval"`
		Format   string        `toml:"date_format"`
	}

	var cfg config

	cfg = config{
		Enable:   true,
		Interval: time.Minute,
		Format:   "Jan _2 2006 (Mon) 3:04 PM",
	}

	decode(decoder, &cfg)

	sendMessage(ch, "Date", cfg.Enable, cfg.Interval, func() json.RawMessage {
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
