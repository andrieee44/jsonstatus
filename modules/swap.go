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
	go sleepMessage(ch, "Swap", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type jsonStruct struct {
			Total, Free int
			FreePerc    float64
		}

		var (
			meminfo             *os.File
			scanner             *bufio.Scanner
			fields              []string
			cached, total, free int
			data                jsonStruct
			dataJSON            json.RawMessage
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

		data = jsonStruct{
			Total:    total,
			Free:     free,
			FreePerc: float64(total-free-cached) / float64(total) * 100,
		}

		dataJSON, err = json.Marshal(data)
		if err != nil {
			panic(err)
		}

		return dataJSON

	})
}
