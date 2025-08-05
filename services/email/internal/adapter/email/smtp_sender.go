package email

import (
	"context"
	"time"

	"email/internal/core/domain"
	"email/internal/core/ports/out"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

type SMTPSender struct {
	client *SMTPClient
	config SMTPConfig
	logger out.Logger
}

func NewSMTPSender(config SMTPConfig, logger out.Logger) out.EmailSender {
	client := NewSMTPClient(config.Host, config.Port, config.User, config.Pass)

	return &SMTPSender{
		client: client,
		config: config,
		logger: logger,
	}
}

func (s *SMTPSender) SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
	s.logger.Infof("Sending email to %s via SMTP", req.To)

	sentAt := time.Now().Unix()

	err := s.client.Send(s.config.User, req.To, req.Subject, []byte(req.Body))
	if err != nil {
		s.logger.Errorf("Failed to send email to %s: %v", req.To, err)
		return &domain.EmailDeliveryResult{
			To:     req.To,
			Status: domain.StatusFailed,
			Error:  err.Error(),
			SentAt: sentAt,
		}, err
	}

	s.logger.Infof("Email sent successfully to %s", req.To)
	return &domain.EmailDeliveryResult{
		To:     req.To,
		Status: domain.StatusDelivered,
		SentAt: sentAt,
	}, nil
}
