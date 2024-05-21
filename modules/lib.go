package modules

import (
	"encoding/json"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Module func(chan<- Message, *os.File)

type Message struct {
	Name string
	Json json.RawMessage
}

func decode(cfgFile *os.File, cfg interface{}) {
	var err error

	_, err = cfgFile.Seek(0, 0)
	if err != nil {
		panic(err)
	}

	_, err = toml.NewDecoder(cfgFile).Decode(cfg)
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
