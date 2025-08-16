//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename email_usecase_mock.go --structname SendEmailUseCase --name SendEmailUseCase
package in

import (
	"context"
	"email/internal/core/domain"
)

type SendEmailUseCase interface {
	SendEmail(ctx context.Context, req domain.SendEmailRequest) (*domain.EmailDeliveryResult, error)
}
