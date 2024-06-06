package modules

import (
	"time"

	"github.com/mafik/pulseaudio"
)

type volConfig struct {
	Enable          bool
	DiscardInterval time.Duration
	Icons           []string
}

func volDiscardUpdates(updates <-chan struct{}, discardInterval time.Duration) {
	var ok bool

	_, ok = <-updates
	IsChanClosed(ok)

	for {
		select {
		case _, ok = <-updates:
			IsChanClosed(ok)
		case <-time.After(discardInterval):
			return
		}
	}
}

func vol(ch chan<- Message, cfg *volConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			client     *pulseaudio.Client
			volume     float32
			volumePerc float64
			mute       bool
			updates    <-chan struct{}
			i          int
			err        error
		)

		for i = 0; i < 10; i++{
			time.Sleep(5)

			client, err = pulseaudio.NewClient()
			if err == nil {
				break
			}
		}

		PanicIf(err)

		updates, err = client.Updates()
		PanicIf(err)

		defer client.Close()

		for {
			volume, err = client.Volume()
			PanicIf(err)

			mute, err = client.Mute()
			PanicIf(err)

			volumePerc = float64(volume) * 100

			sendMessage(ch, "Vol", marshalRawJson(struct {
				Perc float64
				Mute bool
				Icon string
			}{
				Perc: volumePerc,
				Mute: mute,
				Icon: icon(cfg.Icons, 100, volumePerc),
			}))

			volDiscardUpdates(updates, cfg.DiscardInterval)
		}
	}()
}
