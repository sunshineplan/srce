package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/utils/mail"
)

const (
	title   = "SRCE Notification - %s"
	content = "%s\nUser: %s\nIP: %s\n\nCommand: %s"
)

var router = httprouter.New()

func basicAuth(r *http.Request) (bool, string) {
	allowUsers, err := getUsers()
	if err != nil {
		return false, "Error"
	}
	user, password, hasAuth := r.BasicAuth()
	for k, v := range allowUsers {
		if hasAuth && user == k && password == v {
			return true, user
		}
	}
	return false, ""
}

func bash(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	router := strings.Split(strings.Trim(ps.ByName("cmd"), "/ "), "/")
	ip := getClientIP(r)
	allowCommands, commandPath, mailConfig, err := getConfig()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var result string
	switch len(router) {
	case 1:
		for k := range allowCommands {
			if router[0] == k {
				authed, user := basicAuth(r)
				if authed {
					cmd := exec.Command(commandPath + k)
					result, err = run(cmd)
					if err != nil {
						result = fmt.Sprintf("Failed:\n\n%s", err)
					}
					if err := mail.SendMail(
						&mailConfig,
						fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
						fmt.Sprintf(content, time.Now().Format("2006/01/02-15:04:05"), user, ip, cmd),
					); err != nil {
						log.Println(err)
					}
					break
				} else if user == "" {
					w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				} else {
					w.WriteHeader(500)
					return
				}
			}
		}
	case 2:
		for k, v := range allowCommands {
			if router[0] == k {
				for _, arg := range v.([]interface{}) {
					if router[1] == arg {
						authed, user := basicAuth(r)
						if authed {
							cmd := exec.Command(commandPath+k, arg.(string))
							result, err = run(cmd)
							if err != nil {
								result = fmt.Sprintf("Failed:\n\n%s", err)
							}
							if err := mail.SendMail(
								&mailConfig,
								fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
								fmt.Sprintf(content, time.Now().Format("2006/01/02-15:04:05"), user, ip, cmd),
							); err != nil {
								log.Println(err)
							}
							break
						} else if user == "" {
							w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
							http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
							return
						} else {
							w.WriteHeader(500)
							return
						}
					}
				}
			}
		}
	default:
		w.WriteHeader(403)
		return
	}
	w.Write([]byte(result))
}

func forbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
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
