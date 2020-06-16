package misc

import (
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/sunshineplan/metadata"
)

// MetadataConfig is metadata server config
var MetadataConfig = new(metadata.Config)

var (
	// Attempts is default retry attempts
	Attempts = uint(3)
	// Delay is default retry delay
	Delay = 10 * time.Second
	// LastErrorOnly return all errors if false
	LastErrorOnly = true
)

// GetUsers get auth user info
func GetUsers() (users interface{}, err error) {
	err = retry.Do(
		func() (err error) {
			users, err = metadata.Get("srce_user", MetadataConfig)
			return
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay/10),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Failed to get metadata srce_user. #%d: %s\n", n+1, err)
		}),
	)
	return
}

// GetConfig get server setting
func GetConfig() (allowCommands interface{}, commandPath interface{}, mailConfig Subscribe, err error) {
	c := make(chan int)
	go func() {
		err = retry.Do(
			func() (err error) {
				allowCommands, err = metadata.Get("srce_command", MetadataConfig)
				return
			},
			retry.Attempts(Attempts),
			retry.Delay(Delay/10),
			retry.LastErrorOnly(LastErrorOnly),
			retry.OnRetry(func(n uint, err error) {
				log.Printf("Failed to get metadata srce_command. #%d: %s\n", n+1, err)
			}),
		)
		c <- 1
	}()
	go func() {
		err = retry.Do(
			func() (err error) {
				commandPath, err = metadata.Get("srce_path", MetadataConfig)
				return
			},
			retry.Attempts(Attempts),
			retry.Delay(Delay/10),
			retry.LastErrorOnly(LastErrorOnly),
			retry.OnRetry(func(n uint, err error) {
				log.Printf("Failed to get metadata srce_path. #%d: %s\n", n+1, err)
			}),
		)
		c <- 1
	}()

	var srceSubscribe interface{}
	err = retry.Do(
		func() (err error) {
			srceSubscribe, err = metadata.Get("srce_subscribe", MetadataConfig)
			return
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay/10),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Failed to get metadata srce_subscribe. #%d: %s\n", n+1, err)
		}),
	)
	if err == nil {
		mailConfig = Subscribe{
			Sender:         srceSubscribe.(map[string]interface{})["sender"].(string),
			Password:       srceSubscribe.(map[string]interface{})["password"].(string),
			SMTPServer:     srceSubscribe.(map[string]interface{})["smtp_server"].(string),
			SMTPServerPort: int(srceSubscribe.(map[string]interface{})["smtp_server_port"].(float64)),
			Subscriber:     srceSubscribe.(map[string]interface{})["subscriber"].(string),
		}
	}
	_, _ = <-c, <-c
	return
}
