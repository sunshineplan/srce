package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/sunshineplan/utils/mail"
)

type result struct {
	err    error
	stdout []byte
	stderr []byte
}

func basicAuth(r *http.Request) (string, bool) {
	allowUsers, err := getUsers()
	if err != nil {
		return "error", false
	}
	user, password, hasAuth := r.BasicAuth()
	for k, v := range allowUsers {
		if hasAuth && user == k && password == v {
			return user, true
		}
	}
	return "", false
}

func runCmd(cmd *exec.Cmd) (string, error) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	done := make(chan result)
	go func() {
		res := new(result)
		res.stdout, _ = ioutil.ReadAll(stdout)
		res.stderr, _ = ioutil.ReadAll(stderr)
		res.err = cmd.Wait()
		done <- *res
	}()
	select {
	case <-time.After(30 * time.Second):
		return "Process still running.", nil
	case r := <-done:
		return fmt.Sprintf("Output:\n%s\n\nError:\n%s", r.stdout, r.stderr), r.err
	}
}

func execute(user, ip, path, command string, args ...string) string {
	const (
		title   = "SRCE Notification - %s"
		content = "%s\nUser: %s\nIP: %s\n\nCommand: %s"
	)

	cmd := exec.Command(path+command, args...)
	result, err := runCmd(cmd)
	if err != nil {
		result = fmt.Sprintf("Failed:\n\n%s", err)
	}
	if err := (&mail.Dialer{
		Host:     subscribe.SMTPServer,
		Port:     subscribe.SMTPServerPort,
		Account:  subscribe.From,
		Password: subscribe.Password,
	}).Send(&mail.Message{
		To:      subscribe.To,
		Subject: fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
		Body:    fmt.Sprintf(content, time.Now().Format("2006/01/02-15:04:05"), user, ip, cmd),
	},
	); err != nil {
		log.Println(err)
	}
	return result
}

func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
