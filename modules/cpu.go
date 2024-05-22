package modules

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

type cpuConfig struct {
	Enable   bool
	Interval time.Duration
}

func cpuFreq() int {
	var (
		buf  []byte
		freq int
		err  error
	)

	buf, err = os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err != nil {
		panic(err)
	}

	freq, err = strconv.Atoi(string(buf[:len(buf)-1]))
	if err != nil {
		panic(err)
	}

	return freq
}

func cpuAveragePerc(prev []int) float64 {
	return 0
}

func Cpu(ch chan<- Message, cfg *cpuConfig) {
	var prev []int

	go sendMessage(ch, "Cpu", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Frequency   int
			AveragePerc float64
		}

		return marshalRawJson(jsonStruct{
			Frequency:   cpuFreq(),
			AveragePerc: cpuAveragePerc(prev),
		})
	})
}
