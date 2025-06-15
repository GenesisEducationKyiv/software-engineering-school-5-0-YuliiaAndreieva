package email

import (
	"log"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
)

type EmailSender interface {
	SendEmail(to string, subject string, body string) error
}
type emailSender struct {
	host string
	port int
	user string
	pass string
}

func NewEmailSender(host string, port int, user, pass string) EmailSender {
	return &emailSender{host: host, port: port, user: user, pass: pass}
}

func (e *emailSender) SendEmail(to, subject, body string) error {
	log.Printf("Attempting to send email to: %s, subject: %s, from: %s", to, subject, e.user)

	msg := email.NewEmail()
	msg.From = e.user
	msg.To = []string{to}
	msg.Subject = subject
	msg.HTML = []byte(body)

	addr := e.host + ":" + strconv.Itoa(e.port)
	err := msg.Send(addr, smtp.PlainAuth("", e.user, e.pass, e.host))
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Successfully sent email to: %s", to)
	return nil
}
