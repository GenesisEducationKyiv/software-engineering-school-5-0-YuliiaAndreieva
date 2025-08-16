package email

import (
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
)

type SMTPClient struct {
	addr string
	auth smtp.Auth
}

func NewSMTPClient(host string, port int, user, pass string) *SMTPClient {
	addr := host + ":" + strconv.Itoa(port)
	auth := smtp.PlainAuth("", user, pass, host)

	return &SMTPClient{
		addr: addr,
		auth: auth,
	}
}

func (c *SMTPClient) Send(from, to, subject string, htmlBody []byte) error {
	msg := email.NewEmail()
	msg.From = from
	msg.To = []string{to}
	msg.Subject = subject
	msg.HTML = htmlBody

	return msg.Send(c.addr, c.auth)
}
