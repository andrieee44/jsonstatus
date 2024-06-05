package modules

import (
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/fhs/gompd/v2/mpd"
)

type musicConfig struct {
	Enable         bool
	ScrollInterval time.Duration
	Format         string
	Limit          int
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

func musicEvent(watcher *mpd.Watcher, scrollInterval time.Duration, music string, limit, scroll int) (int, bool) {
	var (
		timer    <-chan time.Time
		musicLen int
		ok       bool
		err      error
	)

	musicLen = utf8.RuneCountInString(music)

	if limit != 0 && scrollInterval != 0 && musicLen > limit {
		timer = time.After(scrollInterval)
	}

	select {
	case _, ok = <-watcher.Event:
		IsChanClosed(ok)

		return 0, false
	case err, ok = <-watcher.Error:
		IsChanClosed(ok)
		PanicIf(err)
	case <-timer:
		scroll++

		if scroll > musicLen-limit {
			return 0, true
		}
	}

	return scroll, true
}

func music(ch chan<- Message, cfg *musicConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			watcher      *mpd.Watcher
			music, state string
			scroll        int
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
				Scroll        int
			}{
				Music: music,
				State: state,
				Scroll: scroll,
			}))

			scroll, unchanged = musicEvent(watcher, cfg.ScrollInterval, music, cfg.Limit, scroll)
		}
	}()
}
