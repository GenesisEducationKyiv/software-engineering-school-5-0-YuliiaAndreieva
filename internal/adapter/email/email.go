package email

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/jordan-wright/email"
	"weather-api/internal/core/port"
)

type SMTPEmailSender struct {
	host string
	port int
	user string
	pass string
}

func NewSMTPEmailSender(host string, port int, user, pass string) port.EmailService {
	return &SMTPEmailSender{host: host, port: port, user: user, pass: pass}
}

func (e *SMTPEmailSender) SendEmail(to, subject, body string) error {
	log.Printf("Attempting to send email to: %s, subject: %s, from: %s", to, subject, e.user)

	msg := email.NewEmail()
	msg.From = e.user
	msg.To = []string{to}
	msg.Subject = subject
	msg.HTML = []byte(body)

	err := msg.Send(fmt.Sprintf("%s:%d", e.host, e.port), smtp.PlainAuth("", e.user, e.pass, e.host))
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Successfully sent email to: %s", to)
	return nil
}
