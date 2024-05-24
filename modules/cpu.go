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
	panicIf(err)

	freq, err = strconv.Atoi(string(buf[:len(buf)-1]))
	panicIf(err)

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
	panicIf(err)

	scanner = bufio.NewScanner(stat)
	scanner.Scan()
	panicIf(scanner.Err())

	for i, v = range strings.Fields(scanner.Text())[1:] {
		num, err = strconv.Atoi(v)
		panicIf(err)

		if i == 3 {
			sample.idle = num
		}

		sample.sum += num
	}

	delta = sample.sum - prev.sum
	panicIf(stat.Close())

	return sample, float64(delta-(sample.idle-prev.idle)) / float64(delta) * 100
}

func cpu(ch chan<- Message, cfg *cpuConfig) {
	var prev cpuSample

	go loopMessage(ch, "Cpu", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Frequency   int
			AveragePerc float64
		}

		var perc float64

		prev, perc = cpuAveragePerc(prev)

		return marshalRawJson(jsonStruct{
			Frequency:   cpuFreq(),
			AveragePerc: perc,
		})
	})
}
