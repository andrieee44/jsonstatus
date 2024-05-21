package modules

import (
	"encoding/json"
	"time"

	"github.com/BurntSushi/toml"
)

type Module func(chan<- Message, *toml.Decoder)

type Message struct {
	Name string
	Json json.RawMessage
}

func decode(decoder *toml.Decoder, cfg interface{}) {
	var err error

	_, err = decoder.Decode(cfg)
	if err != nil {
		panic(err)
	}
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
