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

type command struct {
	Path string
	Args []string
}

func getUsers() (users map[string]string, err error) {
	err = meta.Get("srce_user", &users)
	return
}

func getBash() (commands map[string]command, err error) {
	err = meta.Get("srce_command", &commands)
	return
}

func getSubscribe() (subscribe subscribe, err error) {
	err = meta.Get("srce_subscribe", &subscribe)
	return
}
