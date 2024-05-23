package modules

import (
	"encoding/json"
	"time"
)

type ramConfig struct {
	Enable   bool
	Interval time.Duration
}

func Ram(ch chan<- Message, cfg *ramConfig) {
	go loopMessage(ch, "Ram", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free, Available, Used int
			UsedPerc                     float64
		}

		var (
			meminfo map[string]int
			used    int
		)

		meminfo = meminfoMap([]string{
			"MemTotal",
			"MemFree",
			"MemAvailable",
			"Buffers",
			"Cached",
		})

		used = meminfo["MemTotal"] - meminfo["MemFree"] - meminfo["Buffers"] - meminfo["Cached"]

		return marshalRawJson(jsonStruct{
			Total:     meminfo["MemTotal"],
			Free:      meminfo["MemFree"],
			Available: meminfo["MemAvailable"],
			Used:      used,
			UsedPerc:  float64(used) / float64(meminfo["MemTotal"]) * 100,
		})
	})
}
