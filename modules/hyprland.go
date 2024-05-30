package modules

import (
	"encoding/json"
	"errors"
	"net"
	"os"
)

type hyprlandConfig struct {
	Enable bool
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

func hyprlandRequest(sock net.Conn, request string, v any) {
	var (
		err error
	)

	_, err = sock.Write([]byte("-j/" + request))
	panicIf(err)

	panicIf(json.NewDecoder(sock).Decode(v))
}

func hyprlandWindow(sock net.Conn) string {
	type jsonStruct struct {
		Title string
	}

	var window jsonStruct

	window = jsonStruct{}
	hyprlandRequest(sock, "activewindow", &window)

	return window.Title
}

func hyprland(ch chan<- Message, cfg *hyprlandConfig) {
	var (
		sock net.Conn
		err  error
	)

	if !cfg.Enable {
		return
	}

	sock, err = net.Dial("unix", hyprlandSocketsPath()+".socket.sock")
	panicIf(err)

	println(hyprlandWindow(sock))
}
