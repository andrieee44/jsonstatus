package modules

import "github.com/fsnotify/fsnotify"

type briConfig struct {
	Enable bool
}

func bri(ch chan<- Message, cfg *briConfig) {
	const briPath string = "/sys/class/backlight/intel_backlight/brightness"

	var (
		watcher *fsnotify.Watcher
		maxBri  int
	)

	maxBri = pathAtoi("/sys/class/backlight/intel_backlight/max_brightness")

	if !cfg.Enable {
		return
	}

	watcher = mkWatcher([]string{briPath})

	go func() {
		defer func() {
			var err error

			err = watcher.Close()
			if err != nil {
				panic(err)
			}
		}()

		for {
			ch <- Message{
				Name: "Bri",
				Json: marshalRawJson(float64(pathAtoi(briPath)) / float64(maxBri) * 100),
			}

			notifyWatcher(watcher, func(event fsnotify.Event) bool {
				return event.Has(fsnotify.Write)
			})
		}
	}()
}
