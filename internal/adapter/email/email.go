package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
)

type Sender interface {
	SendEmail(opts SendEmailOptions) error
}

type SenderOptions struct {
	Host string
	Port int
	User string
	Pass string
}

type SendEmailOptions struct {
	To      string
	Subject string
	Body    string
}

type emailSender struct {
	host string
	port int
	user string
	pass string
}

func NewSender(opts SenderOptions) Sender {
	return &emailSender{
		host: opts.Host,
		port: opts.Port,
		user: opts.User,
		pass: opts.Pass,
	}
}

func (e *emailSender) SendEmail(opts SendEmailOptions) error {
	log.Printf("Attempting to send email to: %s, subject: %s, from: %s", opts.To, opts.Subject, e.user)

	msg := email.NewEmail()
	msg.From = e.user
	msg.To = []string{opts.To}
	msg.Subject = opts.Subject
	msg.HTML = []byte(opts.Body)

	addr := e.host + ":" + strconv.Itoa(e.port)
	err := msg.Send(addr, smtp.PlainAuth("", e.user, e.pass, e.host))
	if err != nil {
		msg := fmt.Sprintf("unable to send email to %s: %v", opts.To, err)
		log.Print(msg)
		return errors.New(msg)
	}

	log.Printf("Successfully sent email to: %s", opts.To)
	return nil
}
