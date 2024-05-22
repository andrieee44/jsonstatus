package modules

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Date dateConfig
	Ram  ramConfig
	Swap swapConfig
	Cpu  cpuConfig
}

type Message struct {
	Name string
	Json json.RawMessage
}

func DefaultConfig() *Config {
	return &Config{
		Date: dateConfig{
			Enable:   true,
			Interval: time.Minute,
			Format:   "Jan _2 2006 (Mon) 3:04 PM",
		},

		Ram: ramConfig{
			Enable:   true,
			Interval: time.Second,
		},

		Swap: swapConfig{
			Enable:   true,
			Interval: time.Second,
		},

		Cpu: cpuConfig{
			Enable:   true,
			Interval: time.Second,
		},
	}
}

func removeKey(keys []string, key string) ([]string, bool) {
	var (
		i int
		v string
	)

	for i, v = range keys {
		if v == key {
			return append(keys[:i], keys[i+1:]...), true
		}
	}

	return keys, false
}

func meminfoMap(keys []string) map[string]int {
	var (
		keyVal  map[string]int
		meminfo *os.File
		scanner *bufio.Scanner
		fields  []string
		key     string
		val     int
		ok      bool
		err     error
	)

	keyVal = make(map[string]int)

	meminfo, err = os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}

	scanner = bufio.NewScanner(meminfo)

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())
		key = fields[0][:len(fields[0])-1]

		keys, ok = removeKey(keys, key)
		if !ok {
			continue
		}

		val, err = strconv.Atoi(fields[1])
		if err != nil {
			panic(err)
		}

		keyVal[key] = val

		if len(keys) == 0 {
			break
		}
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}

	return keyVal
}

func marshalRawJson(v any) json.RawMessage {
	var (
		data json.RawMessage
		err  error
	)

	data, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

func sendMessage(ch chan<- Message, name string, enable bool, sleep time.Duration, fn func() json.RawMessage) {
	if !enable {
		return
	}

	for {
		ch <- Message{
			Name: name,
			Json: fn(),
		}

		time.Sleep(sleep)
	}
}
