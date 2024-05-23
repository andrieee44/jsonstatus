package modules

import "os/user"

type userConfig struct {
	Enable bool
}

func currentUser(ch chan<- Message, cfg *userConfig) {
	type jsonStruct struct {
		UID, GID, Name string
	}

	var (
		currentUser *user.User
		err         error
	)

	if !cfg.Enable {
		return
	}

	currentUser, err = user.Current()
	if err != nil {
		panic(err)
	}

	go onceMessage(ch, "User", cfg.Enable, marshalRawJson(jsonStruct{
		UID:  currentUser.Uid,
		GID:  currentUser.Gid,
		Name: currentUser.Username,
	}))
}
