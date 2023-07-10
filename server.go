package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/sunshineplan/utils/log"
)

var server = httpsvr.New()

func run() error {
	if *logPath != "" {
		svc.Logger = log.New(*logPath, "", log.LstdFlags)
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

	return server.Run()
}
