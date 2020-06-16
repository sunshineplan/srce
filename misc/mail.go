package misc

import (
	"log"

	"github.com/avast/retry-go"
	"gopkg.in/gomail.v2"
)

// Subscribe info
type Subscribe struct {
	Sender         string
	Password       string
	SMTPServer     string
	SMTPServerPort int
	Subscriber     string
}

// Mail content to subscriber
func Mail(s *Subscribe, subject string, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.Sender)
	m.SetHeader("To", s.Subscriber)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", content)

	d := gomail.NewDialer(s.SMTPServer, s.SMTPServerPort, s.Sender, s.Password)

	err := retry.Do(
		func() error {
			err := d.DialAndSend(m)
			return err
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
		}),
	)
	return err
}
