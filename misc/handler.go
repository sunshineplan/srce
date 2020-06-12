package misc

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	title   = "SRCE Notification - %s"
	content = "%s\nUser: %s\nIP: %s\n\nCommand: %s"
)

// Router default router handler
var Router = httprouter.New()

func init() {
	Router.GET("/bash/*cmd", Bash)
	Router.NotFound = http.HandlerFunc(Forbidden)
}

func basicAuth(r *http.Request) (bool, string) {
	allowUsers := GetUsers()
	user, password, hasAuth := r.BasicAuth()
	for k, v := range allowUsers.(map[string]interface{}) {
		if hasAuth && user == k && password == v {
			return true, user
		}
	}
	return false, ""
}

// Bash handler
func Bash(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	router := strings.Split(strings.Trim(ps.ByName("cmd"), "/ "), "/")
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
	}
	allowCommands, commandPath, mailConfig := GetConfig()
	var result string
	switch len(router) {
	case 1:
		for k := range allowCommands.(map[string]interface{}) {
			if router[0] == k {
				authed, user := basicAuth(r)
				if authed {
					cmd := exec.Command(commandPath.(string) + k)
					result, err = Run(cmd)
					if err != nil {
						result = fmt.Sprintf("Failed:\n\n%s", err)
					}
					if err := Mail(
						&mailConfig,
						fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
						fmt.Sprintf(content, time.Now().Format("2006/01/02-15:04:05"), user, ip, cmd),
					); err != nil {
						log.Print(err)
					}
					break
				} else {
					w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
			}
		}
	case 2:
		for k, v := range allowCommands.(map[string]interface{}) {
			if router[0] == k {
				for _, arg := range v.([]interface{}) {
					if router[1] == arg {
						authed, user := basicAuth(r)
						if authed {
							cmd := exec.Command(commandPath.(string)+k, arg.(string))
							result, err = Run(cmd)
							if err != nil {
								result = fmt.Sprintf("Failed:\n\n%s", err)
							}
							if err := Mail(
								&mailConfig,
								fmt.Sprintf(title, time.Now().Format("20060102 15:04:05")),
								fmt.Sprintf(content, time.Now().Format("2006/01/02-15:04:05"), user, ip, cmd),
							); err != nil {
								log.Print(err)
							}
							break
						} else {
							w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
							http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
							return
						}
					}
				}
			}
		}
	default:
		result = ""
	}
	if result == "" {
		w.WriteHeader(403)
	} else {
		w.Write([]byte(result))
	}
}

// Forbidden all other router
func Forbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
}
