package modules

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type uptimeConfig struct {
	Enable   bool
	Interval time.Duration
}

func uptime(ch chan<- Message, cfg *uptimeConfig) {
	go loopMessage(ch, "Uptime", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Hours, Minutes, Seconds int
		}

		var (
			buf       []byte
			uptime    float64
			uptimeInt int
			err       error
		)

		buf, err = os.ReadFile("/proc/uptime")
		if err != nil {
			panic(err)
		}

		uptime, err = strconv.ParseFloat(strings.Fields(string(buf))[0], 64)
		if err != nil {
			panic(err)
		}

		uptimeInt = int(uptime)

		return marshalRawJson(jsonStruct{
			Hours:   uptimeInt / 3600,
			Minutes: (uptimeInt % 3600) / 60,
			Seconds: uptimeInt % 60,
		})
	})
}
