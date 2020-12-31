package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func run() {
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	router := httprouter.New()
	server.Handler = router

	router.GET("/bash/*cmd", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		router := strings.Split(strings.Trim(ps.ByName("cmd"), "/ "), "/")
		ip = getClientIP(r)
		commands, path, err := getConfig()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		user = ""
		var authed bool
		var result string
		switch len(router) {
		case 1:
			for k := range commands {
				if router[0] == k {
					authed, user = basicAuth(r)
					if authed {
						result = execute(path, router[0])
						w.Write([]byte(result))
						return
					}
				}
			}
		case 2:
			for k, v := range commands {
				if router[0] == k {
					for _, arg := range v {
						if router[1] == arg {
							authed, user = basicAuth(r)
							if authed {
								result = execute(path, router[0], router[1])
								w.Write([]byte(result))
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
		if user == "" {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		w.WriteHeader(500)
	})

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
