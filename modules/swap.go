package modules

import (
	"encoding/json"
	"time"
)

type swapConfig struct {
	Enable   bool
	Interval time.Duration
	Icons    []string
}

func swap(ch chan<- Message, cfg *swapConfig) {
	go loopMessage(ch, "Swap", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free, Used int
			UsedPerc          float64
			Icon              string
		}

		var (
			meminfo                   map[string]int
			total, free, cached, used int
			usedPerc                  float64
		)

		meminfo = meminfoMap([]string{
			"SwapCached",
			"SwapTotal",
			"SwapFree",
		})

		total, free, cached = meminfo["SwapTotal"], meminfo["SwapFree"], meminfo["SwapCached"]
		used = total - free + cached
		usedPerc = float64(used) / float64(total) * 100

		return marshalRawJson(jsonStruct{
			Total:    total,
			Free:     free,
			Used:     used,
			UsedPerc: usedPerc,
			Icon:     icon(cfg.Icons, 100, usedPerc),
		})
	})
}
