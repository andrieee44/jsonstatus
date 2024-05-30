package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type batConfig struct {
	Enable   bool
	Interval time.Duration
	Icons    []string
}

func bat(ch chan<- Message, cfg *batConfig) {
	go loopMessage(ch, "Bat", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type batInfo struct {
			Name, Status, Icon string
			Capacity           int
		}

		var (
			bats     []batInfo
			batPaths []string
			v        string
			buf      []byte
			capacity int
			err      error
		)

		batPaths, err = filepath.Glob("/sys/class/power_supply/BAT*")
		PanicIf(err)

		for _, v = range batPaths {
			buf, err = os.ReadFile(v + "/status")
			PanicIf(err)

			capacity = pathAtoi(v + "/capacity")

			bats = append(bats, batInfo{
				Name:     filepath.Base(v),
				Status:   string(buf[:len(buf)-1]),
				Icon:     icon(cfg.Icons, 100, float64(capacity)),
				Capacity: capacity,
			})
		}

		return marshalRawJson(bats)
	})
}
