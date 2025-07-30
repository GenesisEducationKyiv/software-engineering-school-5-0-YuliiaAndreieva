package in

import (
	"context"
	"email/internal/core/domain"
)

type SendEmailUseCase interface {
	SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error)
	SendWeatherUpdateEmail(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error)
}
