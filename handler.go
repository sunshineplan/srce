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

func shell(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ip, admin, ok := auth(w, r)
	if !ok {
		return
	}

	cmd, err := parseCmd(ps.ByName("cmd"))
	if err != nil {
		log.Print(err)
		w.WriteHeader(400)
		return
	}

	if !admin {
		log.Printf("%s has no permission to run shell: %s", user, cmd)
		w.WriteHeader(403)
		return
	}

	result := execute(user, ip, "", cmd[0], cmd[1:]...)
	w.Write([]byte(result))
}

func cmd(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ip, _, ok := auth(w, r)
	if !ok {
		return
	}

	cmd, err := parseCmd(ps.ByName("cmd"))
	if err != nil {
		log.Print(err)
		w.WriteHeader(400)
		return
	}

	commands, err := getCmd()
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	switch len(cmd) {
	case 1:
		for k, v := range commands {
			if cmd[0] == k {
				result := execute(user, ip, v.Path, cmd[0])
				w.Write([]byte(result))
				return
			}
		}
	case 2:
		for k, v := range commands {
			if cmd[0] == k {
				for _, arg := range v.Args {
					if cmd[1] == arg {
						result := execute(user, ip, v.Path, cmd[0], cmd[1])
						w.Write([]byte(result))
						return
					}
				}
			}
		}
	default:
		w.WriteHeader(400)
	}
}

func email(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, ip, _, ok := auth(w, r)
	if !ok {
		return
	}

	var data struct {
		S, B, T string
		A       []struct{ F, D string }
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Print(err)
		w.WriteHeader(400)
		return
	}

	key := r.Header.Get("X-Key")
	subject, err := cipher.DecryptText(key, data.S)
	if err != nil {
		log.Print(err)
		subject = "Unknow"
	}
	body, err := cipher.DecryptText(key, data.B)
	if err != nil {
		log.Print(err)
		body = "Unknow"
	}
	var to mail.Receipts
	if data.T != "" {
		toStr, err := cipher.DecryptText(key, data.T)
		if err != nil {
			log.Print(err)
		}
		to, err = mail.ParseReceipts(toStr)
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			return
		}
	}
	var attachments []*mail.Attachment
	for i, a := range data.A {
		filename, err := cipher.DecryptText(key, a.F)
		if err != nil {
			filename = "Unknow" + strconv.Itoa(i)
		}
		b, err := base64.StdEncoding.DecodeString(a.D)
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			return
		}
		b, err = cipher.Decrypt([]byte(key), b)
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			return
		}
		attachments = append(attachments, &mail.Attachment{Filename: filename, Bytes: b})
	}

	dialer, subscriber, err := getSubscribe()
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}
	if len(to) == 0 {
		to = subscriber
	}
	if err := dialer.Send(&mail.Message{
		To:          to,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	}); err != nil {
		log.Print(err)
		w.WriteHeader(502)
	} else {
		log.Printf("SRCE Mail Sent - User: %s, IP: %s, Subject: %s, To: %s", user, ip, subject, to)
	}
}

func crypto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key := r.FormValue("key")
	content := r.FormValue("content")
	switch r.FormValue("mode") {
	case "encrypt":
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(map[string]any{"result": cipher.EncryptText(key, content)})
		w.Write(data)
	case "decrypt":
		w.Header().Set("Content-Type", "application/json")
		result, err := cipher.DecryptText(key, strings.TrimSpace(content))
		var data []byte
		if err != nil {
			data, _ = json.Marshal(map[string]any{"result": nil})
		} else {
			data, _ = json.Marshal(map[string]any{"result": result})
		}
		w.Write(data)
	default:
		w.WriteHeader(400)
	}
}
