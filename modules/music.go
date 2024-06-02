package modules

import (
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/fhs/gompd/v2/mpd"
)

type musicConfig struct {
	Enable   bool
	Interval time.Duration
	Format   string
	Limit    int
}

func musicGet(format string) (string, string) {
	var (
		client        *mpd.Client
		music, status mpd.Attrs
		err           error
	)

	client, err = mpd.Dial("tcp", "127.0.0.1:6600")
	PanicIf(err)

	music, err = client.CurrentSong()
	PanicIf(err)

	status, err = client.Status()
	PanicIf(err)

	PanicIf(client.Close())

	return regexp.MustCompilePOSIX("%[A-Za-z]+%").ReplaceAllStringFunc(format, func(str string) string {
		return music[str[1:len(str)-1]]
	}), status["state"]
}

func musicEvent(watcher *mpd.Watcher, interval time.Duration, music string, limit, index int) (int, bool) {
	var (
		timer    <-chan time.Time
		musicLen int
		ok       bool
		err      error
	)

	musicLen = utf8.RuneCountInString(music)

	if limit != 0 && interval != 0 && musicLen > limit {
		timer = time.After(interval)
	}

	select {
	case _, ok = <-watcher.Event:
		IsChanClosed(ok)

		return 0, false
	case err, ok = <-watcher.Error:
		IsChanClosed(ok)
		PanicIf(err)
	case <-timer:
		index++

		if index > musicLen-limit {
			return 0, true
		}
	}

	return index, true
}

func music(ch chan<- Message, cfg *musicConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			watcher      *mpd.Watcher
			music, state string
			index        int
			unchanged    bool
			err          error
		)

		watcher, err = mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
		PanicIf(err)

		defer func() {
			PanicIf(watcher.Close())
		}()

		for {
			if !unchanged {
				music, state = musicGet(cfg.Format)
			}

			sendMessage(ch, "Music", marshalRawJson(struct {
				Music, State string
				Index        int
			}{
				Music: music,
				State: state,
				Index: index,
			}))

			index, unchanged = musicEvent(watcher, cfg.Interval, music, cfg.Limit, index)
		}
	}()
}
