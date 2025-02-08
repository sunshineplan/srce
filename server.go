package main

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/sunshineplan/utils/httpsvr"
)

var server = httpsvr.New()

func run() error {
	if err := os.MkdirAll(*uploadPath, 0750); err != nil {
		return err
	}

	router := httprouter.New()
	server.Handler = router

	router.GET("/shell/*cmd", shell)
	router.GET("/cmd/*cmd", cmd)
	router.POST("/upload", upload)
	router.POST("/mail", email)
	router.POST("/crypto", crypto)
	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	} {
		router.Handle(method, "/test", testHTTPRequest)
	}

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})

	return server.Run()
}
