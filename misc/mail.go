package misc

import (
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

	return d.DialAndSend(m)
}
