package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/vharitonsky/iniflags"
)

// OS is the running program's operating system
const OS = runtime.GOOS

func main() {
	flag.StringVar(&metadataConfig.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&metadataConfig.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&metadataConfig.VerifyValue, "value", "", "Verify Header Value")
	unix := flag.String("unix", "", "UNIX-domain Socket")
	host := flag.String("host", "127.0.0.1", "Server Host")
	port := flag.String("port", "12345", "Server Port")
	logPath := flag.String("log", "", "Log Path")
	iniflags.SetConfigFile("config.ini")
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	router.GET("/bash/*cmd", bash)
	router.NotFound = http.HandlerFunc(forbidden)

	if *unix != "" && OS == "linux" {
		if _, err := os.Stat(*unix); err == nil {
			err = os.Remove(*unix)
			if err != nil {
				log.Fatalf("Failed to remove socket file: %v", err)
			}
		}

		listener, err := net.Listen("unix", *unix)
		if err != nil {
			log.Fatalf("Failed to listen socket file: %v", err)
		}

		idleConnsClosed := make(chan struct{})
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			if err := listener.Close(); err != nil {
				log.Printf("Failed to close listener: %v", err)
			}
			if _, err := os.Stat(*unix); err == nil {
				if err := os.Remove(*unix); err != nil {
					log.Printf("Failed to remove socket file: %v", err)
				}
			}
			close(idleConnsClosed)
		}()

		if err := os.Chmod(*unix, 0666); err != nil {
			log.Fatalf("Failed to chmod socket file: %v", err)
		}

		http.Serve(listener, router)
		<-idleConnsClosed
	} else {
		http.ListenAndServe(*host+":"+*port, router)
	}
}
