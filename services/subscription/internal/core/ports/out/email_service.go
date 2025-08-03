package out

import (
	"context"
	"subscription/internal/core/domain"
)

type EmailService interface {
	SendConfirmationEmail(ctx context.Context, request domain.ConfirmationEmailRequest) error
}

type EmailDeliveryResult struct {
	EmailID string
	To      string
	Status  string
	Error   string
	SentAt  int64
}
