package email

import (
	"context"
	"net/smtp"
	"strconv"
	"time"

	"email-service/internal/core/domain"
	"email-service/internal/core/ports/out"

	"github.com/jordan-wright/email"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

type SMTPSender struct {
	config SMTPConfig
	logger out.Logger
}

func NewSMTPSender(config SMTPConfig, logger out.Logger) out.EmailSender {
	return &SMTPSender{
		config: config,
		logger: logger,
	}
}

func (s *SMTPSender) SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
	s.logger.Infof("Sending email to %s via SMTP", req.To)
	
	msg := email.NewEmail()
	msg.From = s.config.User
	msg.To = []string{req.To}
	msg.Subject = req.Subject
	msg.HTML = []byte(req.Body)

	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	err := msg.Send(addr, smtp.PlainAuth("", s.config.User, s.config.Pass, s.config.Host))
	if err != nil {
		s.logger.Errorf("Failed to send email to %s: %v", req.To, err)
		return &domain.EmailDeliveryResult{
			To:     req.To,
			Status: domain.StatusFailed,
			Error:  err.Error(),
			SentAt: time.Now().Unix(),
		}, err
	}

	s.logger.Infof("Email sent successfully to %s", req.To)
	return &domain.EmailDeliveryResult{
		To:     req.To,
		Status: domain.StatusDelivered,
		SentAt: time.Now().Unix(),
	}, nil
} 