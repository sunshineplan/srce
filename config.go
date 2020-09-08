package main

import (
	"encoding/json"
	"sync"

	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

var metadataConfig metadata.Config

func getUsers() (map[string]string, error) {
	b, err := metadataConfig.Get("srce_user")
	if err != nil {
		return nil, err
	}
	var users map[string]string
	if err := json.Unmarshal(b, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func getConfig() (command map[string][]string, path string, subscribe mail.Setting, err error) {
	var wg sync.WaitGroup
	done := make(chan error, 3*2)
	wg.Add(3)
	go func() {
		defer wg.Done()
		b, err := metadataConfig.Get("srce_command")
		done <- err
		done <- json.Unmarshal(b, &command)
	}()
	go func() {
		defer wg.Done()
		b, err := metadataConfig.Get("srce_path")
		done <- err
		done <- json.Unmarshal(b, &path)
	}()
	go func() {
		defer wg.Done()
		b, err := metadataConfig.Get("srce_subscribe")
		done <- err
		done <- json.Unmarshal(b, &subscribe)
	}()
	wg.Wait()
	close(done)
	for e := range done {
		if e != nil {
			err = e
			return
		}
	}
	return
}
