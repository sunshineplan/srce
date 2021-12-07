package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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

func basicAuth(r *http.Request) (string, bool, bool) {
	allowUsers, err := getUsers()
	if err != nil {
		return "error", false, false
	}
	user, password, hasAuth := r.BasicAuth()
	for name, info := range allowUsers {
		if hasAuth && user == name && password == info.Password {
			return user, info.Admin, true
		}
	}
	return "", false, false
}

func auth(w http.ResponseWriter, r *http.Request) (user, ip string, admin, ok bool) {
	ip = getClientIP(r)
	user, admin, ok = basicAuth(r)
	if !ok {
		if user == "" {
			w.Header().Set("WWW-Authenticate", "Basic realm=SRCE")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		w.WriteHeader(500)
	}
	return
}

func parseCmd(cmd string) (res []string, err error) {
	cmd, err = url.QueryUnescape(strings.Trim(cmd, "/ "))
	if err != nil {
		return
	}
	cmd = strings.ReplaceAll(cmd, "_", " ")

	for _, s := range strings.Split(cmd, "/") {
		res = append(res, strings.Fields(strings.ReplaceAll(s, "~", "/"))...)
	}
	return
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
		res.stdout, _ = io.ReadAll(stdout)
		res.stderr, _ = io.ReadAll(stderr)
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
	subscribe, err := getSubscribe()
	if err != nil {
		log.Print(err)
		return result
	}
	if err := (&mail.Dialer{
		Host:     subscribe.SMTPServer,
		Port:     subscribe.SMTPServerPort,
		Account:  subscribe.From,
		Password: subscribe.Password,
	}).Send(&mail.Message{
		To:      subscribe.To,
		Subject: fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
		Body:    fmt.Sprintf(content, time.Now().Format("2006/01/02 - 15:04:05"), user, ip, cmd),
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
