package modules

import (
	"os"
	"os/user"
)

type userConfig struct {
	Enable bool
}

func currentUser(ch chan<- Message, cfg *userConfig) {
	var (
		currentUser *user.User
		host        string
		err         error
	)

	if !cfg.Enable {
		return
	}

	currentUser, err = user.Current()
	PanicIf(err)

	host, err = os.Hostname()
	PanicIf(err)

	go onceMessage(ch, "User", cfg.Enable, marshalRawJson(struct {
		UID, GID, Name, Host string
	}{
		UID:  currentUser.Uid,
		GID:  currentUser.Gid,
		Name: currentUser.Username,
		Host: host,
	}))
}
