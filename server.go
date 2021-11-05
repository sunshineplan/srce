package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/cipher"
	"github.com/sunshineplan/utils/mail"
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
		ip := getClientIP(r)
		commands, path, err := getConfig()
		if err != nil {
			log.Print(err)
			w.WriteHeader(500)
			return
		}

		var user string
		var ok bool
		switch len(router) {
		case 1:
			for k := range commands {
				if router[0] == k {
					user, ok = basicAuth(r)
					if ok {
						result := execute(user, ip, path, router[0])
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
							user, ok = basicAuth(r)
							if ok {
								result := execute(user, ip, path, router[0], router[1])
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
			w.Header().Set("WWW-Authenticate", "Basic realm=SRCE")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		w.WriteHeader(500)
	})

	router.POST("/mail", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ip := getClientIP(r)
		user, ok := basicAuth(r)
		if !ok {
			if user == "" {
				w.Header().Set("WWW-Authenticate", "Basic realm=SRCE")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			w.WriteHeader(500)
			return
		}
		if r == nil {
			w.WriteHeader(500)
			return
		}
		var data struct {
			T, B string
			A    []struct{ N, D string }
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&data); err != nil {
			w.WriteHeader(400)
			return
		}

		key := r.Header.Get("x-key")
		title, err := cipher.DecryptText(key, data.T)
		if err != nil {
			title = "Unknow"
		}
		body, err := cipher.DecryptText(key, data.B)
		if err != nil {
			body = "Unknow"
		}
		var attachments []*mail.Attachment
		for i, a := range data.A {
			name, err := cipher.DecryptText(key, a.N)
			if err != nil {
				name = "Unknow" + strconv.Itoa(i)
			}
			data, err := cipher.Decrypt([]byte(key), []byte(a.D))
			if err != nil {
				w.WriteHeader(400)
				return
			}
			attachments = append(attachments, &mail.Attachment{Filename: name, Bytes: data})
		}

		if err := (&mail.Dialer{
			Host:     subscribe.SMTPServer,
			Port:     subscribe.SMTPServerPort,
			Account:  subscribe.From,
			Password: subscribe.Password,
		}).Send(&mail.Message{
			To:          subscribe.To,
			Subject:     title + " " + ip,
			Body:        body,
			Attachments: attachments,
		},
		); err != nil {
			log.Println(err)
		}
	})

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
