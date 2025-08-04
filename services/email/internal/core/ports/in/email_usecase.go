//go:generate mockery --dir . --output ../../../../tests/mocks --outpkg mocks --filename email_usecase_mock.go --structname SendEmailUseCase --name SendEmailUseCase
package in

import (
	"context"
	"email/internal/core/domain"
)

type SendEmailUseCase interface {
	SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error)
	SendWeatherUpdateEmail(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error)
}
