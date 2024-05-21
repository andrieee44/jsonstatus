package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/andrieee44/jsonstatus/modules"
)

func configDir() string {
	const dirname string = "/jsonfetch"

	var dir string

	dir = os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return dir + dirname
	}

	dir = os.Getenv("HOME")
	if dir != "" {
		return dir + "/.config" + dirname
	}

	panic(errors.New("$HOME is empty"))
}

func configFile() *os.File {
	var (
		file *os.File
		dir  string
		err  error
	)

	if len(os.Args) == 2 {
		file, err = os.OpenFile(os.Args[1], os.O_RDONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}

		return file
	}

	dir = configDir()

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	file, err = os.OpenFile(dir+"/jsonfetch.toml", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	return file
}

func main() {
	var (
		ch       chan modules.Message
		msgMap   map[string]json.RawMessage
		cfgFile  *os.File
		msg      modules.Message
		jsonData []byte
		err      error
	)

	ch = make(chan modules.Message)
	msgMap = make(map[string]json.RawMessage)
	cfgFile = configFile()

	modules.Date(ch, cfgFile)
	modules.Ram(ch, cfgFile)

	for msg = range ch {
		msgMap[msg.Name] = msg.Json

		jsonData, err = json.Marshal(msgMap)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(jsonData))
	}
}
