package modules

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

type hyprlandConfig struct {
	Enable         bool
	ScrollInterval time.Duration
	Limit          int
}

type hyprlandWorkspace struct {
	Id   int
	Name string
}

type hyprlandMonitor struct {
	Active     int
	Workspaces []hyprlandWorkspace
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

	return filepath.Join(runtime, "hypr", his)
}

func hyprlandEventChan(path string) (<-chan string, net.Conn) {
	var (
		events     net.Conn
		eventsChan chan string
		scanner    *bufio.Scanner
		event      string
		err        error
	)

	events, err = net.Dial("unix", filepath.Join(path, ".socket2.sock"))
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
			case strings.HasPrefix(event, "destroyworkspace>>"):
			default:
				continue
			}

			eventsChan <- event
		}
	}()

	return eventsChan, events
}

func hyprlandEvent(eventsChan <-chan string, scrollInterval time.Duration, window string, limit, scroll int) (int, bool) {
	var (
		timer     <-chan time.Time
		windowLen int
		ok        bool
	)

	windowLen = utf8.RuneCountInString(window)

	if limit != 0 && scrollInterval != 0 && windowLen > limit {
		timer = time.After(scrollInterval)
	}

	select {
	case _, ok = <-eventsChan:
		PanicIfClosed(ok)

		return 0, false
	case _, ok = <-timer:
		PanicIfClosed(ok)

		scroll++
		if scroll > windowLen-limit {
			scroll = 0
		}

		return scroll, true
	}
}

func hyprlandRequest(path string) (string, map[string]*hyprlandMonitor) {
	type workspace struct {
		Id            int
		Monitor, Name string
	}

	type monitor struct {
		Name string

		ActiveWorkspace struct {
			Id int
		}
	}

	var (
		query             net.Conn
		decoder           *json.Decoder
		monitors          []monitor
		mon               monitor
		workspaces        []workspace
		ws                workspace
		monitorWorkspaces map[string]*hyprlandMonitor
		err               error

		win struct {
			Title string
		}
	)

	query, err = net.Dial("unix", filepath.Join(path, ".socket.sock"))
	PanicIf(err)

	_, err = query.Write([]byte("[[BATCH]]j/activewindow;j/monitors;j/workspaces"))
	PanicIf(err)

	decoder = json.NewDecoder(query)
	PanicIf(decoder.Decode(&win))
	PanicIf(decoder.Decode(&monitors))
	PanicIf(decoder.Decode(&workspaces))
	PanicIf(query.Close())

	monitorWorkspaces = make(map[string]*hyprlandMonitor)

	for _, mon = range monitors {
		monitorWorkspaces[mon.Name] = &hyprlandMonitor{
			Active: mon.ActiveWorkspace.Id,
		}
	}

	for _, ws = range workspaces {
		monitorWorkspaces[ws.Monitor].Workspaces = append(monitorWorkspaces[ws.Monitor].Workspaces, hyprlandWorkspace{
			Id:   ws.Id,
			Name: ws.Name,
		})
	}

	return win.Title, monitorWorkspaces
}

func hyprland(ch chan<- Message, cfg *hyprlandConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			path, window      string
			events            net.Conn
			eventsChan        <-chan string
			monitorWorkspaces map[string]*hyprlandMonitor
			scroll            int
			unchanged         bool
		)

		path = hyprlandSocketsPath()
		eventsChan, events = hyprlandEventChan(path)

		defer func() {
			PanicIf(events.Close())
		}()

		for {
			if !unchanged {
				window, monitorWorkspaces = hyprlandRequest(path)
			}

			sendMessage(ch, "Hyprland", marshalRawJson(struct {
				Window        string
				Monitors      map[string]*hyprlandMonitor
				Scroll, Limit int
			}{
				Window:   window,
				Monitors: monitorWorkspaces,
				Scroll:   scroll,
				Limit:    cfg.Limit,
			}))

			scroll, unchanged = hyprlandEvent(eventsChan, cfg.ScrollInterval, window, cfg.Limit, scroll)
		}
	}()
}
