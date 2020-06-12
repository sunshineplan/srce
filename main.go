package main

import (
	"flag"
	"log"
	"net/http"

	"srce-go/misc"

	"github.com/vharitonsky/iniflags"
)

func main() {
	flag.StringVar(&misc.MetadataConfig.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&misc.MetadataConfig.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&misc.MetadataConfig.VerifyValue, "value", "", "Verify Header Value")
	host := flag.String("host", "127.0.0.1", "Server Host")
	port := flag.String("port", "12345", "Server Port")
	iniflags.SetConfigFile("config.ini")
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	log.Fatal(http.ListenAndServe(*host+":"+*port, misc.Router))
}
