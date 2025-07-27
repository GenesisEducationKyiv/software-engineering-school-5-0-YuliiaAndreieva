package in

import (
	"context"
	"email-service/internal/core/domain"
)

// EmailService defines the contract for email operations
type EmailService interface {
	SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error)
	SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error)
	SendWeatherUpdateEmail(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error)
}

// EmailBuilderService defines the contract for email template building
type EmailBuilderService interface {
	BuildEmail(ctx context.Context, req domain.EmailBuilderRequest) (*domain.EmailBuilderResponse, error)
	BuildConfirmationEmail(ctx context.Context, city, token, baseURL string) (*domain.EmailBuilderResponse, error)
	BuildWeatherUpdateEmail(ctx context.Context, city string, temperature float64, humidity int, description, token, baseURL string) (*domain.EmailBuilderResponse, error)
} 