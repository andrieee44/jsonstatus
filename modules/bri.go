package modules

import "github.com/fsnotify/fsnotify"

type briConfig struct {
	Enable bool
	Icons  []string
}

func bri(ch chan<- Message, cfg *briConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		const briPath string = "/sys/class/backlight/intel_backlight/brightness"

		var (
			watcher *fsnotify.Watcher
			maxBri  int
			perc    float64
		)

		maxBri = pathAtoi("/sys/class/backlight/intel_backlight/max_brightness")
		watcher = mkWatcher([]string{briPath})

		defer func() {
			PanicIf(watcher.Close())
		}()

		for {
			perc = float64(pathAtoi(briPath)) / float64(maxBri) * 100

			sendMessage(ch, "Bri", marshalRawJson(struct {
				Perc float64
				Icon string
			}{
				Perc: perc,
				Icon: icon(cfg.Icons, 100, perc),
			}))

			notifyWatcher(watcher, func(event fsnotify.Event) bool {
				return event.Has(fsnotify.Write)
			})
		}
	}()
}
