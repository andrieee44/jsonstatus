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
}

func Bat(ch chan<- Message, cfg *batConfig) {
	go sendMessage(ch, "Bat", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type batInfo struct {
			Status   string
			Capacity int
		}

		var (
			bats     map[string]batInfo
			batPaths []string
			v        string
			buf      []byte
			err      error
		)

		bats = make(map[string]batInfo)

		batPaths, err = filepath.Glob("/sys/class/power_supply/BAT*")
		if err != nil {
			panic(err)
		}

		for _, v = range batPaths {
			buf, err = os.ReadFile(v + "/status")
			if err != nil {
				panic(err)
			}

			bats[filepath.Base(v)] = batInfo{
				Status:   string(buf[:len(buf)-1]),
				Capacity: pathAtoi(v + "/capacity"),
			}
		}

		return marshalRawJson(bats)
	})
}
