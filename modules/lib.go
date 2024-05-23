package modules

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Date  dateConfig
	Ram   ramConfig
	Swap  swapConfig
	Cpu   cpuConfig
	Bri   briConfig
	Bat   batConfig
	Music musicConfig
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

		Bri: briConfig{
			Enable: true,
		},

		Bat: batConfig{
			Enable:   true,
			Interval: time.Minute,
		},

		Music: musicConfig{
			Enable:   true,
			Interval: time.Second,
		},
	}
}

func mkWatcher(files []string) *fsnotify.Watcher {
	var (
		watcher *fsnotify.Watcher
		v       string
		err     error
	)

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, v = range files {
		err = watcher.Add(v)
		if err != nil {
			panic(err)
		}
	}

	return watcher
}

func notifyWatcher(watcher *fsnotify.Watcher, handler func(fsnotify.Event) bool) {
	var (
		event fsnotify.Event
		ok    bool
		err   error
	)

	for {
		select {
		case event, ok = <-watcher.Events:
			if !ok || handler(event) {
				return
			}
		case err, ok = <-watcher.Errors:
			if !ok {
				return
			}

			if err != nil {
				panic(err)
			}
		}
	}
}

func pathAtoi(path string) int {
	var (
		buf []byte
		num int
		err error
	)

	buf, err = os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	num, err = strconv.Atoi(string(buf[:len(buf)-1]))
	if err != nil {
		panic(err)
	}

	return num
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

	err = meminfo.Close()
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
