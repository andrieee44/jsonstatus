package modules

import (
	"regexp"
	"time"

	"github.com/fhs/gompd/v2/mpd"
)

type musicConfig struct {
	Enable   bool
	Interval time.Duration
	Format   string
}

func musicFmt(regex *regexp.Regexp, music mpd.Attrs, format string) string {
	return regex.ReplaceAllStringFunc(format, func(str string) string {
		return music[str[1:len(str)-1]]
	})
}

func music(ch chan<- Message, cfg *musicConfig) {
	var (
		regex         *regexp.Regexp
		client        *mpd.Client
		watcher       *mpd.Watcher
		music, status mpd.Attrs
		musicStr      string
		index         int
		ok            bool
		err           error
	)

	if !cfg.Enable {
		return
	}

	regex = regexp.MustCompilePOSIX("%[A-Za-z]+%")

	client, err = mpd.Dial("tcp", "127.0.0.1:6600")
	PanicIf(err)

	watcher, err = mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
	PanicIf(err)

	go func() {
		type jsonStruct struct {
			Music, State string
			Index        int
		}

		defer func() {
			PanicIf(client.Close())
			PanicIf(watcher.Close())
		}()

		for {
			music, err = client.CurrentSong()
			PanicIf(err)

			status, err = client.Status()
			PanicIf(err)

			musicStr = musicFmt(regex, music, cfg.Format)

			ch <- Message{
				Name: "Music",
				Json: marshalRawJson(jsonStruct{
					State: status["state"],
					Music: musicStr,
					Index: index,
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

				PanicIf(err)
			case <-time.After(cfg.Interval):
				index++

				if index >= len(musicStr) {
					index = 0
				}
			}
		}
	}()
}
