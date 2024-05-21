package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func Ram(ch chan<- Message, cfgFile *os.File) {
	type config struct {
		Enable   bool          `toml:"ram_enable"`
		Interval time.Duration `toml:"ram_interval"`
	}

	var cfg config

	cfg = config{
		Enable:   true,
		Interval: time.Second,
	}

	decode(cfgFile, &cfg)

	go sendMessage(ch, "Ram", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free, Available, Used int
			UsedPerc                     float64
		}

		var (
			meminfo                                       *os.File
			total, free, available, buffers, cached, used int
			data                                          jsonStruct
			dataJSON                                      json.RawMessage
			err                                           error
		)

		meminfo, err = os.Open("/proc/meminfo")
		if err != nil {
			panic(err)
		}

		defer meminfo.Close()

		_, err = fmt.Fscanf(meminfo, `MemTotal: %d kB
MemFree: %d kB
MemAvailable: %d kB
Buffers: %d kB
Cached: %d kB
`, &total, &free, &available, &buffers, &cached)
		if err != nil {
			panic(err)
		}

		used = total - free - buffers - cached

		data = jsonStruct{
			Total:     total,
			Free:      free,
			Available: available,
			Used:      used,
			UsedPerc:  float64(used) / float64(total) * 100,
		}

		dataJSON, err = json.Marshal(data)
		if err != nil {
			panic(err)
		}

		return dataJSON
	})
}
