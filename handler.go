package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/cipher"
	"github.com/sunshineplan/utils/mail"
)

func bash(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	router := strings.Split(strings.Trim(ps.ByName("cmd"), "/ "), "/")
	ip := getClientIP(r)
	commands, path, err := getBash()
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
}

func email(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		A    []struct{ F, D string }
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
		filename, err := cipher.DecryptText(key, a.F)
		if err != nil {
			filename = "Unknow" + strconv.Itoa(i)
		}
		b, err := base64.StdEncoding.DecodeString(a.D)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		b, err = cipher.Decrypt([]byte(key), b)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		attachments = append(attachments, &mail.Attachment{Filename: filename, Bytes: b})
	}

	subscribe, err := getSubscribe()
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}
	if err := (&mail.Dialer{
		Host:     subscribe.SMTPServer,
		Port:     subscribe.SMTPServerPort,
		Account:  subscribe.From,
		Password: subscribe.Password,
	}).Send(&mail.Message{
		To:          subscribe.To,
		Subject:     title,
		Body:        body,
		Attachments: attachments,
	},
	); err != nil {
		log.Println(err)
		w.WriteHeader(502)
	} else {
		log.Printf("SRCE Mail Sent - User: %s, IP: %s, Title: %s", user, ip, title)
	}
}

func crypto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type M map[string]interface{}
	mode := r.FormValue("mode")
	key := r.FormValue("key")
	content := r.FormValue("content")
	switch mode {
	case "encrypt":
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(M{"result": cipher.EncryptText(key, content)})
		w.Write(data)
	case "decrypt":
		w.Header().Set("Content-Type", "application/json")
		result, err := cipher.DecryptText(key, strings.TrimSpace(content))
		var data []byte
		if err != nil {
			data, _ = json.Marshal(M{"result": nil})
		} else {
			data, _ = json.Marshal(M{"result": result})
		}
		w.Write(data)
	default:
		w.WriteHeader(400)
	}
}
