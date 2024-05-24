package modules

import (
	"time"

	"github.com/fhs/gompd/v2/mpd"
)

type musicConfig struct {
	Enable   bool
	Interval time.Duration
}

func music(ch chan<- Message, cfg *musicConfig) {
	type jsonStruct struct {
		Music, Status mpd.Attrs
		Index         int
	}

	var (
		client        *mpd.Client
		watcher       *mpd.Watcher
		music, status mpd.Attrs
		index         int
		ok            bool
		err           error
	)

	if !cfg.Enable {
		return
	}

	client, err = mpd.Dial("tcp", "127.0.0.1:6600")
	panicIf(err)

	watcher, err = mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
	panicIf(err)

	go func() {
		defer func() {
			panicIf(client.Close())
			panicIf(watcher.Close())
		}()

		for {
			music, err = client.CurrentSong()
			panicIf(err)

			status, err = client.Status()
			panicIf(err)

			ch <- Message{
				Name: "Music",
				Json: marshalRawJson(jsonStruct{
					Music:  music,
					Status: status,
					Index:  index,
				}),
			}

			select {
			case _, ok = <-watcher.Event:
				if !ok {
					return
				}

				index = 0
			case err, ok = <-watcher.Error:
				if !ok {
					return
				}

				panicIf(err)
			case <-time.After(cfg.Interval):
				index++

				if index < 0 {
					index = 0
				}
			}
		}
	}()
}
