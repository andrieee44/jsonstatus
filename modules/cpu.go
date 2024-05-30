package modules

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type cpuConfig struct {
	Enable   bool
	Interval time.Duration
	Icons    []string
}

type cpuSample struct {
	sum, idle int
}

func cpuFreq() int {
	var (
		buf  []byte
		freq int
		err  error
	)

	buf, err = os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	PanicIf(err)

	freq, err = strconv.Atoi(string(buf[:len(buf)-1]))
	PanicIf(err)

	return freq
}

func cpuAveragePerc(prev cpuSample) (cpuSample, float64) {
	var (
		stat          *os.File
		scanner       *bufio.Scanner
		v             string
		i, num, delta int
		sample        cpuSample
		err           error
	)

	stat, err = os.Open("/proc/stat")
	PanicIf(err)

	scanner = bufio.NewScanner(stat)
	scanner.Scan()
	PanicIf(scanner.Err())

	for i, v = range strings.Fields(scanner.Text())[1:] {
		num, err = strconv.Atoi(v)
		PanicIf(err)

		if i == 3 {
			sample.idle = num
		}

		sample.sum += num
	}

	delta = sample.sum - prev.sum
	PanicIf(stat.Close())

	return sample, float64(delta-(sample.idle-prev.idle)) / float64(delta) * 100
}

func cpu(ch chan<- Message, cfg *cpuConfig) {
	var prev cpuSample

	go loopMessage(ch, "Cpu", cfg.Enable, cfg.Interval, func() json.RawMessage {
		var perc float64

		prev, perc = cpuAveragePerc(prev)

		return marshalRawJson(struct {
			Frequency   int
			AveragePerc float64
			Icon        string
		}{
			Frequency:   cpuFreq(),
			AveragePerc: perc,
			Icon:        icon(cfg.Icons, 100, perc),
		})
	})
}
