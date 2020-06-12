package misc

import "log"

// GetUsers get auth user info
func GetUsers() interface{} {
	users, err := Metadata("srce_user")
	if err != nil {
		log.Print(err)
	}
	return users
}

// GetConfig get server setting
func GetConfig() (allowCommands interface{}, commandPath interface{}, mailConfig Subscribe) {
	allowCommands, err := Metadata("srce_command")
	if err != nil {
		log.Print(err)
	}
	commandPath, err = Metadata("srce_path")
	if err != nil {
		log.Print(err)
	}

	metadata, err := Metadata("srce_subscribe")
	if err != nil {
		log.Print(err)
	}
	mailConfig = Subscribe{
		Sender:         metadata.(map[string]interface{})["sender"].(string),
		Password:       metadata.(map[string]interface{})["password"].(string),
		SMTPServer:     metadata.(map[string]interface{})["smtp_server"].(string),
		SMTPServerPort: int(metadata.(map[string]interface{})["smtp_server_port"].(float64)),
		Subscriber:     metadata.(map[string]interface{})["subscriber"].(string),
	}
	return
}
