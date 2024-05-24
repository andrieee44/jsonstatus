package modules

import (
	"time"

	"github.com/mafik/pulseaudio"
)

type volConfig struct {
	Enable bool
}

func vol(ch chan<- Message, cfg *volConfig) {
	type jsonStruct struct {
		Volume float64
		Mute   bool
	}

	var (
		client   *pulseaudio.Client
		volume   float32
		mute, ok bool
		updates  <-chan struct{}
		err      error
	)

	if !cfg.Enable {
		return
	}

	client, err = pulseaudio.NewClient()
	if err != nil {
		panic(err)
	}

	updates, err = client.Updates()
	if err != nil {
		panic(err)
	}

	go func() {
		defer client.Close()

		for {
			volume, err = client.Volume()
			if err != nil {
				panic(err)
			}

			mute, err = client.Mute()
			if err != nil {
				panic(err)
			}

			ch <- Message{
				Name: "Vol",
				Json: marshalRawJson(jsonStruct{
					Volume: float64(volume) * 100,
					Mute:   mute,
				}),
			}

			_, ok = <-updates
			if !ok {
				return
			}

		DISCARD:
			for {
				select {
				case _, ok = <-updates:
					if !ok {
						return
					}
				case <-time.After(time.Millisecond * 10):
					break DISCARD
				}
			}
		}
	}()
}
