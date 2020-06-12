package misc

import (
	"log"

	"github.com/sunshineplan/metadata"
)

// MetadataConfig is metadata server config
var MetadataConfig = new(metadata.Config)

// GetUsers get auth user info
func GetUsers() interface{} {
	users, err := metadata.Get("srce_user", MetadataConfig)
	if err != nil {
		log.Print(err)
	}
	return users
}

// GetConfig get server setting
func GetConfig() (allowCommands interface{}, commandPath interface{}, mailConfig Subscribe) {
	allowCommands, err := metadata.Get("srce_command", MetadataConfig)
	if err != nil {
		log.Print(err)
	}
	commandPath, err = metadata.Get("srce_path", MetadataConfig)
	if err != nil {
		log.Print(err)
	}

	srceSubscribe, err := metadata.Get("srce_subscribe", MetadataConfig)
	if err != nil {
		log.Print(err)
	}
	mailConfig = Subscribe{
		Sender:         srceSubscribe.(map[string]interface{})["sender"].(string),
		Password:       srceSubscribe.(map[string]interface{})["password"].(string),
		SMTPServer:     srceSubscribe.(map[string]interface{})["smtp_server"].(string),
		SMTPServerPort: int(srceSubscribe.(map[string]interface{})["smtp_server_port"].(float64)),
		Subscriber:     srceSubscribe.(map[string]interface{})["subscriber"].(string),
	}
	return
}
