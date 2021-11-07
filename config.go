package main

import (
	"github.com/sunshineplan/utils/metadata"
)

var meta metadata.Server

type subscribe struct {
	From, SMTPServer, Password string
	SMTPServerPort             int
	To                         []string
}

func getUsers() (users map[string]string, err error) {
	err = meta.Get("srce_user", &users)
	return
}

func getBash() (command map[string][]string, path string, err error) {
	c := make(chan error, 1)
	go func() {
		c <- meta.Get("srce_command", &command)
	}()

	var data struct{ Path string }
	if err = meta.Get("srce_path", &data); err != nil {
		return
	}
	path = data.Path

	err = <-c

	return
}

func getSubscribe() (subscribe subscribe, err error) {
	err = meta.Get("srce_subscribe", &subscribe)
	return
}
