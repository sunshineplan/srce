package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/vharitonsky/iniflags"
)

var server = httpsvr.New()
var logPath string

var svc = service.Service{
	Name:    "SRCE",
	Desc:    "Instance to serve Simple Remote Command Execution",
	Exec:    run,
	Options: service.Options{Dependencies: []string{"After=network.target"}},
}

var (
	joinPath = filepath.Join
	dir      = filepath.Dir
)

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "127.0.0.1", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	exclude := flag.String("exclude", "", "Exclude Files")
	flag.StringVar(&logPath, "log", "", "Log Path")
	iniflags.SetConfigFile(joinPath(dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.SetAllowUnknownFlags(true)
	iniflags.Parse()

	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run":
			svc.Run(false)
		case "debug":
			svc.Run(true)
		case "install":
			err = svc.Install()
		case "remove":
			err = svc.Remove()
		case "start":
			err = svc.Start()
		case "stop":
			err = svc.Stop()
		case "restart":
			err = svc.Restart()
		case "update":
			err = svc.Update()
		default:
			log.Fatalln("Unknown argument:", flag.Arg(0))
		}
	default:
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}
	if err != nil {
		log.Fatalf("failed to %s: %v", flag.Arg(0), err)
	}
}
