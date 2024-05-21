package modules

import (
	"encoding/json"
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
