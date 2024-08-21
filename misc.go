package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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
		var res result
		res.stdout, _ = io.ReadAll(stdout)
		res.stderr, _ = io.ReadAll(stderr)
		res.err = cmd.Wait()
		done <- res
	}()
	select {
	case <-time.After(30 * time.Second):
		return "Process still running.", nil
	case r := <-done:
		return fmt.Sprintf("Output:\n%s\n\nError:\n%s", r.stdout, r.stderr), r.err
	}
}

func execute(user, ip, path, command string, args ...string) (res string) {
	cmd := exec.Command(path+command, args...)
	res, err := runCmd(cmd)
	if err != nil {
		res = fmt.Sprintf("%s\n\nFailed:\n%s", res, err)
	}
	svc.Printf("SRCE Execute - User: %s IP: %s Command: %s", user, ip, cmd)
	dialer, to, err := getSubscribe()
	if err != nil {
		svc.Print(err)
		return
	}
	if err := dialer.Send(&mail.Message{
		To:      to,
		Subject: fmt.Sprintf("SRCE Notification - %s", time.Now().Format("20060102 15:04:05")),
		Body:    fmt.Sprintf("User: %s\nIP: %s\n\nCommand: %s", user, ip, cmd),
	}); err != nil {
		svc.Print(err)
	}
	return
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(filepath.Join(*uploadPath, dst))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
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
