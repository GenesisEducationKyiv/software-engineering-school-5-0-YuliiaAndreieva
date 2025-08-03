package out

import (
	"context"
	"email/internal/core/domain"
)

type EmailSender interface {
	SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error)
}
