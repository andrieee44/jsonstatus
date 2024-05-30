package modules

import (
	"time"

	"github.com/mafik/pulseaudio"
)

type volConfig struct {
	Enable  bool
	Discard time.Duration
	Icons   []string
}

func vol(ch chan<- Message, cfg *volConfig) {
	type jsonStruct struct {
		Volume float64
		Mute   bool
		Icon   string
	}

	var (
		client     *pulseaudio.Client
		volume     float32
		volumePerc float64
		mute, ok   bool
		updates    <-chan struct{}
		err        error
	)

	if !cfg.Enable {
		return
	}

	client, err = pulseaudio.NewClient()
	panicIf(err)

	updates, err = client.Updates()
	panicIf(err)

	go func() {
		defer client.Close()

		for {
			volume, err = client.Volume()
			panicIf(err)

			mute, err = client.Mute()
			panicIf(err)

			volumePerc = float64(volume) * 100

			ch <- Message{
				Name: "Vol",
				Json: marshalRawJson(jsonStruct{
					Volume: volumePerc,
					Mute:   mute,
					Icon:   icon(cfg.Icons, 100, volumePerc),
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
				case <-time.After(cfg.Discard):
					break DISCARD
				}
			}
		}
	}()
}
