package modules

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type swapConfig struct {
	Enable   bool
	Interval time.Duration
}

func Swap(ch chan<- Message, cfg *swapConfig) {
	go sendMessage(ch, "Swap", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free int
			UsedPerc    float64
		}

		var (
			meminfo             *os.File
			scanner             *bufio.Scanner
			fields              []string
			cached, total, free int
			err                 error
		)

		meminfo, err = os.Open("/proc/meminfo")
		if err != nil {
			panic(err)
		}

		defer meminfo.Close()

		scanner = bufio.NewScanner(meminfo)

		for {
			if !scanner.Scan() {
				err = scanner.Err()
				if err != nil {
					panic(err)
				}

				break
			}

			fields = strings.Fields(scanner.Text())

			switch fields[0] {
			case "SwapCached:":
				cached, err = strconv.Atoi(fields[1])
				if err != nil {
					panic(err)
				}

			case "SwapTotal:":
				total, err = strconv.Atoi(fields[1])
				if err != nil {
					panic(err)
				}

			case "SwapFree:":
				free, err = strconv.Atoi(fields[1])
				if err != nil {
					panic(err)
				}

				break
			}
		}

		return marshalRawJson(jsonStruct{
			Total:    total,
			Free:     free,
			UsedPerc: float64(total-cached-free) / float64(total) * 100,
		})
	})
}
