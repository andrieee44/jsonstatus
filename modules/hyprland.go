package modules

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"time"
)

type hyprlandConfig struct {
	Enable   bool
	Interval time.Duration
	Limit    int
}

type hyprlandWorkspace struct {
	Id   int
	Name string
}

func hyprlandSocketsPath() string {
	var his, runtime string

	his = os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		panic(errors.New("HYPRLAND_INSTANCE_SIGNATURE is empty"))
	}

	runtime = os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		panic(errors.New("XDG_RUNTIME_DIR is empty"))
	}

	return runtime + "/hypr/" + his + "/"
}

func hyprlandEventChan(path string) (<-chan string, net.Conn) {
	var (
		events     net.Conn
		eventsChan chan string
		scanner    *bufio.Scanner
		event      string
		err        error
	)

	events, err = net.Dial("unix", path+".socket2.sock")
	PanicIf(err)

	eventsChan = make(chan string)
	scanner = bufio.NewScanner(events)

	go func() {
		for {
			if !scanner.Scan() {
				PanicIf(scanner.Err())
				return
			}

			event = scanner.Text()

			switch {
			case strings.HasPrefix(event, "workspace>>"):
			case strings.HasPrefix(event, "activewindow>>"):
			default:
				continue
			}

			eventsChan <- event
		}
	}()

	return eventsChan, events
}

func hyprlandEvent(eventsChan <-chan string, interval time.Duration, window string, limit, index int) (int, bool) {
	var (
		timer <-chan time.Time
		ok    bool
	)

	if limit != 0 && interval != 0 && len(window) > limit {
		timer = time.After(interval)
	}

	for {
		select {
		case _, ok = <-eventsChan:
			IsChanClosed(ok)

			return 0, false
		case <-timer:
			index++

			if index > len(window)-limit {
				index = 0
			}

			return index, true
		}
	}
}

func hyprlandRequest(path string) (string, []hyprlandWorkspace, int) {
	type window struct {
		Title string
	}

	type monitor struct {
		ActiveWorkspace struct {
			Id int
		}
	}

	var (
		query      net.Conn
		decoder    *json.Decoder
		win        window
		workspaces []hyprlandWorkspace
		monitors   []monitor
		err        error
	)

	query, err = net.Dial("unix", path+".socket.sock")
	PanicIf(err)

	_, err = query.Write([]byte("[[BATCH]]j/activewindow;j/workspaces;j/monitors"))
	PanicIf(err)

	decoder = json.NewDecoder(query)
	PanicIf(decoder.Decode(&win))
	PanicIf(decoder.Decode(&workspaces))
	PanicIf(decoder.Decode(&monitors))

	PanicIf(query.Close())

	return win.Title, workspaces, monitors[0].ActiveWorkspace.Id
}

func hyprland(ch chan<- Message, cfg *hyprlandConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			path, window  string
			events        net.Conn
			eventsChan    <-chan string
			workspaces    []hyprlandWorkspace
			active, index int
			unchanged     bool
		)

		path = hyprlandSocketsPath()
		eventsChan, events = hyprlandEventChan(path)

		defer func() {
			PanicIf(events.Close())
		}()

		for {
			if !unchanged {
				window, workspaces, active = hyprlandRequest(path)
			}

			sendMessage(ch, "Hyprland", marshalRawJson(struct {
				Window        string
				Workspaces    []hyprlandWorkspace
				Active, Index int
			}{
				Window:     window,
				Workspaces: workspaces,
				Active:     active,
				Index:      index,
			}))

			index, unchanged = hyprlandEvent(eventsChan, cfg.Interval, window, cfg.Limit, index)
		}
	}()
}
