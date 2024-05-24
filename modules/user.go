package modules

import (
	"os"
	"os/user"
)

type userConfig struct {
	Enable bool
}

func currentUser(ch chan<- Message, cfg *userConfig) {
	type jsonStruct struct {
		UID, GID, Name, Host string
	}

	var (
		currentUser *user.User
		host        string
		err         error
	)

	if !cfg.Enable {
		return
	}

	currentUser, err = user.Current()
	if err != nil {
		panic(err)
	}

	host, err = os.Hostname()
	if err != nil {
		panic(err)
	}

	go onceMessage(ch, "User", cfg.Enable, marshalRawJson(jsonStruct{
		UID:  currentUser.Uid,
		GID:  currentUser.Gid,
		Name: currentUser.Username,
		Host: host,
	}))
}
