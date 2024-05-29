package modules

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Date   dateConfig
	Ram    ramConfig
	Swap   swapConfig
	Cpu    cpuConfig
	Bri    briConfig
	Bat    batConfig
	Music  musicConfig
	Vol    volConfig
	Uptime uptimeConfig
	User   userConfig
	Disk   diskConfig
}

type Message struct {
	Name string
	Json json.RawMessage
}

func Run(ch chan<- Message, cfg *Config) {
	date(ch, &cfg.Date)
	ram(ch, &cfg.Ram)
	swap(ch, &cfg.Swap)
	cpu(ch, &cfg.Cpu)
	bri(ch, &cfg.Bri)
	bat(ch, &cfg.Bat)
	music(ch, &cfg.Music)
	vol(ch, &cfg.Vol)
	uptime(ch, &cfg.Uptime)
	currentUser(ch, &cfg.User)
	disk(ch, &cfg.Disk)
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
			Format:   "%AlbumArtist% - %Title%",
		},

		Vol: volConfig{
			Enable: true,
		},

		Uptime: uptimeConfig{
			Enable:   true,
			Interval: time.Minute,
		},

		User: userConfig{
			Enable: true,
		},

		Disk: diskConfig{
			Enable:   true,
			Interval: time.Minute,
			Disks:    []string{"/"},
		},
	}
}

func panicIf(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func mkWatcher(files []string) *fsnotify.Watcher {
	var (
		watcher *fsnotify.Watcher
		v       string
		err     error
	)

	watcher, err = fsnotify.NewWatcher()
	panicIf(err)

	for _, v = range files {
		panicIf(watcher.Add(v))
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

			panicIf(err)
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
	panicIf(err)

	num, err = strconv.Atoi(string(buf[:len(buf)-1]))
	panicIf(err)

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
	panicIf(err)

	scanner = bufio.NewScanner(meminfo)

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())
		key = fields[0][:len(fields[0])-1]

		keys, ok = removeKey(keys, key)
		if !ok {
			continue
		}

		val, err = strconv.Atoi(fields[1])
		panicIf(err)

		keyVal[key] = val

		if len(keys) == 0 {
			break
		}
	}

	panicIf(scanner.Err())
	panicIf(meminfo.Close())

	return keyVal
}

func marshalRawJson(v any) json.RawMessage {
	var (
		data json.RawMessage
		err  error
	)

	data, err = json.Marshal(v)
	panicIf(err)

	return data
}

func sendMessage(ch chan<- Message, name string, msg json.RawMessage) {
	ch <- Message{
		Name: name,
		Json: msg,
	}
}

func onceMessage(ch chan<- Message, name string, enable bool, msg json.RawMessage) {
	if !enable {
		return
	}

	sendMessage(ch, name, msg)
}

func loopMessage(ch chan<- Message, name string, enable bool, sleep time.Duration, fn func() json.RawMessage) {
	if !enable {
		return
	}

	for {
		sendMessage(ch, name, fn())
		time.Sleep(sleep)
	}
}
