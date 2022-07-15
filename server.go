package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/utils/httpsvr"
)

var server = httpsvr.New()

func run() {
	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	router := httprouter.New()
	server.Handler = router

	router.GET("/shell/*cmd", shell)
	router.GET("/cmd/*cmd", cmd)
	router.POST("/mail", email)
	router.POST("/crypto", crypto)

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
