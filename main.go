package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/andrieee44/jsonstatus/modules"
)

func configDir() string {
	const dirname string = "jsonstatus"

	var dir string

	dir = os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return filepath.Join(dir, dirname)
	}

	dir = os.Getenv("HOME")
	if dir != "" {
		return filepath.Join(dir, ".config", dirname)
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
		modules.PanicIf(err)

		return file
	}

	dir = configDir()
	modules.PanicIf(os.MkdirAll(dir, 0755))

	file, err = os.OpenFile(filepath.Join(dir, "jsonstatus.toml"), os.O_RDONLY|os.O_CREATE, 0644)
	modules.PanicIf(err)

	return file
}

func configToml() *modules.Config {
	var (
		cfgFile *os.File
		cfg     *modules.Config
		err     error
	)

	cfgFile = configFile()
	cfg = modules.DefaultConfig()

	_, err = toml.NewDecoder(cfgFile).Decode(cfg)
	modules.PanicIf(err)

	modules.PanicIf(cfgFile.Close())

	return cfg
}

func main() {
	var (
		ch       chan modules.Message
		messages map[string]json.RawMessage
		message  modules.Message
		data     []byte
		err      error
	)

	ch = make(chan modules.Message)
	messages = make(map[string]json.RawMessage)
	modules.Run(ch, configToml())

	for message = range ch {
		messages[message.Name] = message.Data

		data, err = json.Marshal(messages)
		modules.PanicIf(err)

		fmt.Println(string(data))
	}
}
