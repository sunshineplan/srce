package main

import (
	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

var meta metadata.Server

type command struct {
	Path string
	Args []string
}

type info struct {
	Password string
	Admin    bool
}

func getUsers() (users map[string]info, err error) {
	err = meta.Get("srce_user", &users)
	return
}

func getCmd() (commands map[string]command, err error) {
	err = meta.Get("srce_command", &commands)
	return
}

func getSubscribe() (*mail.Dialer, mail.Receipts, error) {
	var subscribe struct {
		From, SMTPServer, Password string
		SMTPServerPort             int
		To                         mail.Receipts
	}
	if err := meta.Get("srce_subscribe", &subscribe); err != nil {
		return nil, nil, err
	}

	return &mail.Dialer{
		Server:   subscribe.SMTPServer,
		Port:     subscribe.SMTPServerPort,
		Account:  subscribe.From,
		Password: subscribe.Password,
	}, subscribe.To, nil
}
