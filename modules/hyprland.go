package modules

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
)

type hyprlandConfig struct {
	Enable bool
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

func hyprlandEvent(scanner *bufio.Scanner) {
	if !scanner.Scan() {
		PanicIf(scanner.Err())
	}
}

func hyprlandRequest(request string, v any) {
	var (
		query net.Conn
		err   error
	)

	query, err = net.Dial("unix", hyprlandSocketsPath()+".socket.sock")
	PanicIf(err)

	_, err = query.Write([]byte("-j/" + request))
	PanicIf(err)

	PanicIf(json.NewDecoder(query).Decode(v))
	PanicIf(query.Close())
}

func hyprlandWindow() string {
	type window struct {
		Title string
	}

	var win window

	win = window{}
	hyprlandRequest("activewindow", &win)

	return win.Title
}

func hyprlandWorkspaces() []hyprlandWorkspace {
	var workspaces []hyprlandWorkspace

	hyprlandRequest("workspaces", &workspaces)

	return workspaces
}

func hyprlandActive() int {
	type monitor struct {
		ActiveWorkspace struct {
			Id int
		}
	}

	var monitors []monitor

	hyprlandRequest("monitors", &monitors)

	return monitors[0].ActiveWorkspace.Id
}

func hyprland(ch chan<- Message, cfg *hyprlandConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			events  net.Conn
			scanner *bufio.Scanner
			err     error
		)

		events, err = net.Dial("unix", hyprlandSocketsPath()+".socket2.sock")
		PanicIf(err)

		scanner = bufio.NewScanner(events)

		defer func() {
			PanicIf(events.Close())
		}()

		for {
			sendMessage(ch, "Hyprland", marshalRawJson(struct {
				Active     int
				Workspaces []hyprlandWorkspace
				Window     string
			}{
				Active:     hyprlandActive(),
				Workspaces: hyprlandWorkspaces(),
				Window:     hyprlandWindow(),
			}))

			hyprlandEvent(scanner)
		}
	}()
}
