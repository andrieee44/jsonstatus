package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ramConfig struct {
	Enable   bool
	Interval time.Duration
}

func Ram(ch chan<- Message, cfg *ramConfig) {
	go sendMessage(ch, "Ram", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free, Available, Used int
			UsedPerc                     float64
		}

		var (
			meminfo                                       *os.File
			total, free, available, buffers, cached, used int
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

		return marshalRawJson(jsonStruct{
			Total:     total,
			Free:      free,
			Available: available,
			Used:      used,
			UsedPerc:  float64(used) / float64(total) * 100,
		})
	})
}
