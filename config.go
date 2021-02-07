package main

import (
	"sync"

	"github.com/sunshineplan/utils/metadata"
)

var meta metadata.Server

var subscribe struct {
	From, SMTPServer, Password string
	SMTPServerPort             int
	To                         []string
}

func getUsers() (users map[string]string, err error) {
	err = meta.Get("srce_user", &users)
	return
}

func getConfig() (command map[string][]string, path string, err error) {
	var wg sync.WaitGroup
	done := make(chan error, 3)
	wg.Add(3)
	go func() {
		defer wg.Done()
		done <- meta.Get("srce_command", &command)
	}()
	go func() {
		defer wg.Done()
		var p struct{ Path string }
		if err := meta.Get("srce_path", &p); err != nil {
			done <- err
		}
		path = p.Path
	}()
	go func() {
		defer wg.Done()
		done <- meta.Get("srce_subscribe", &subscribe)
	}()
	wg.Wait()
	close(done)
	for err = range done {
		if err != nil {
			return
		}
	}
	return
}
