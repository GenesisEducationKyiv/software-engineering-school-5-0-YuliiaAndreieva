package out

type SendEmailOptions struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	SendEmail(opts SendEmailOptions) error
}
