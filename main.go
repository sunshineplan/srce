package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/txt"
)

var svc = service.New()

func init() {
	svc.Name = "SRCE"
	svc.Desc = "Instance to serve Simple Remote Command Execution"
	svc.Exec = run
	svc.Options = service.Options{
		Dependencies: []string{"Wants=network-online.target", "After=network.target"},
		ExcludeFiles: []string{"scripts/srce.conf"},
	}
}

var (
	maxMemory  = flag.Int64("max", 32, "Max Memory(MB)")
	uploadPath = flag.String("upload", "upload", "Upload Path")
	logPath    = flag.String("log", "", "Log File Path")
	envPath    = flag.String("env", ".env", "Environment File Path")
)

func main() {
	self, err := os.Executable()
	if err != nil {
		svc.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "127.0.0.1", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	flags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	flags.Parse()

	if *logPath != "" {
		svc.SetLogger(*logPath, "", log.LstdFlags)
	}

	if s, err := txt.ReadFile(*envPath); err == nil {
		for _, i := range s {
			if s := strings.SplitN(i, "=", 2); len(s) == 2 {
				os.Setenv(strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
			} else if s := strings.TrimSpace(i); s != "" {
				svc.Println("Unknown env:", s)
			}
		}
	} else if *envPath != "" {
		svc.Println("Failed to load env file:", err)
	}

	if err := svc.ParseAndRun(flag.Args()); err != nil {
		svc.Fatal(err)
	}
}
