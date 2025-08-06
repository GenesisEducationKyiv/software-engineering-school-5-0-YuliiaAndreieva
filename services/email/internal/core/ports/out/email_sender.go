//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename email_sender_mock.go --structname EmailSender --name EmailSender
package out

import (
	"context"
	"email/internal/core/domain"
)

type EmailSender interface {
	SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error)
}
