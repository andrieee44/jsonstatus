package modules

import (
	"encoding/json"
	"time"
)

type ramConfig struct {
	Enable   bool
	Interval time.Duration
	Icons    []string
}

func ram(ch chan<- Message, cfg *ramConfig) {
	go loopMessage(ch, "Ram", cfg.Enable, cfg.Interval, func() json.RawMessage {
		var (
			meminfo  map[string]int
			used     int
			usedPerc float64
		)

		meminfo = meminfoMap([]string{
			"MemTotal",
			"MemFree",
			"MemAvailable",
			"Buffers",
			"Cached",
		})

		used = meminfo["MemTotal"] - meminfo["MemFree"] - meminfo["Buffers"] - meminfo["Cached"]
		usedPerc = float64(used) / float64(meminfo["MemTotal"]) * 100

		return marshalRawJson(struct {
			Total, Free, Available, Used int
			UsedPerc                     float64
			Icon                         string
		}{
			Total:     meminfo["MemTotal"],
			Free:      meminfo["MemFree"],
			Available: meminfo["MemAvailable"],
			Used:      used,
			UsedPerc:  usedPerc,
			Icon:      icon(cfg.Icons, 100, usedPerc),
		})
	})
}
