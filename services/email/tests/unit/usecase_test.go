package unit

import (
	"context"
	"errors"
	"testing"

	"email/internal/adapter/logger"
	"email/internal/core/domain"
	"email/internal/core/usecase"
	"github.com/stretchr/testify/assert"
)

type MockEmailSender struct {
	sendEmailFunc func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error)
}

func (m *MockEmailSender) SendEmail(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, req)
	}
	return &domain.EmailDeliveryResult{
		To:     req.To,
		Status: domain.StatusDelivered,
	}, nil
}

type MockTemplateBuilder struct {
	buildConfirmationEmailFunc  func(ctx context.Context, email, city, confirmationLink string) (string, error)
	buildWeatherUpdateEmailFunc func(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error)
}

func (m *MockTemplateBuilder) BuildConfirmationEmail(ctx context.Context, email, city, confirmationLink string) (string, error) {
	if m.buildConfirmationEmailFunc != nil {
		return m.buildConfirmationEmailFunc(ctx, email, city, confirmationLink)
	}
	return "<html><body>Test confirmation email</body></html>", nil
}

func (m *MockTemplateBuilder) BuildWeatherUpdateEmail(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error) {
	if m.buildWeatherUpdateEmailFunc != nil {
		return m.buildWeatherUpdateEmailFunc(ctx, email, city, description, humidity, windSpeed, temperature)
	}
	return "<html><body>Test weather update email</body></html>", nil
}

func TestSendEmailUseCase_SendConfirmationEmail_Success(t *testing.T) {
	logger := logger.NewLogrusLogger()

	tests := []struct {
		name           string
		request        domain.ConfirmationEmailRequest
		expectedStatus domain.EmailDeliveryStatus
	}{
		{
			name: "Successful confirmation email",
			request: domain.ConfirmationEmailRequest{
				To:               "test@example.com",
				Subject:          "Confirm Subscription",
				City:             "Kyiv",
				ConfirmationLink: "http://localhost/confirm/token123",
			},
			expectedStatus: domain.StatusDelivered,
		},
		{
			name: "Another valid confirmation email",
			request: domain.ConfirmationEmailRequest{
				To:               "user@test.com",
				Subject:          "Confirm Weather Subscription",
				City:             "Lviv",
				ConfirmationLink: "http://localhost/confirm/token456",
			},
			expectedStatus: domain.StatusDelivered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateBuilder := &MockTemplateBuilder{
				buildConfirmationEmailFunc: func(ctx context.Context, email, city, confirmationLink string) (string, error) {
					return "<html><body>Test template</body></html>", nil
				},
			}

			emailSender := &MockEmailSender{
				sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
					return &domain.EmailDeliveryResult{
						To:     req.To,
						Status: domain.StatusDelivered,
					}, nil
				},
			}

			useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

			result, err := useCase.SendConfirmationEmail(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
		})
	}
}

func TestSendEmailUseCase_SendConfirmationEmail_TemplateBuilderError(t *testing.T) {
	logger := logger.NewLogrusLogger()

	request := domain.ConfirmationEmailRequest{
		To:               "test@example.com",
		Subject:          "Confirm Subscription",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token123",
	}

	templateBuilder := &MockTemplateBuilder{
		buildConfirmationEmailFunc: func(ctx context.Context, email, city, confirmationLink string) (string, error) {
			return "", errors.New("template builder error")
		},
	}

	emailSender := &MockEmailSender{
		sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendConfirmationEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "template builder error")
}

func TestSendEmailUseCase_SendConfirmationEmail_EmailSenderError(t *testing.T) {
	logger := logger.NewLogrusLogger()

	request := domain.ConfirmationEmailRequest{
		To:               "test@example.com",
		Subject:          "Confirm Subscription",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token123",
	}

	templateBuilder := &MockTemplateBuilder{
		buildConfirmationEmailFunc: func(ctx context.Context, email, city, confirmationLink string) (string, error) {
			return "<html><body>Test template</body></html>", nil
		},
	}

	emailSender := &MockEmailSender{
		sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
			return nil, errors.New("email sender error")
		},
	}

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendConfirmationEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email sender error")
}

func TestSendEmailUseCase_SendWeatherUpdateEmail_Success(t *testing.T) {
	logger := logger.NewLogrusLogger()

	tests := []struct {
		name           string
		request        domain.WeatherUpdateEmailRequest
		expectedStatus domain.EmailDeliveryStatus
	}{
		{
			name: "Successful weather update email",
			request: domain.WeatherUpdateEmailRequest{
				To:          "test@example.com",
				Subject:     "Weather Update",
				Name:        "User",
				City:        "Kyiv",
				Temperature: 15,
				Description: "Partly cloudy",
				Humidity:    65,
				WindSpeed:   12,
			},
			expectedStatus: domain.StatusDelivered,
		},
		{
			name: "Another valid weather update email",
			request: domain.WeatherUpdateEmailRequest{
				To:          "user@test.com",
				Subject:     "Daily Weather Report",
				Name:        "John",
				City:        "Lviv",
				Temperature: 22,
				Description: "Sunny",
				Humidity:    45,
				WindSpeed:   8,
			},
			expectedStatus: domain.StatusDelivered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateBuilder := &MockTemplateBuilder{
				buildWeatherUpdateEmailFunc: func(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error) {
					return "<html><body>Test weather template</body></html>", nil
				},
			}

			emailSender := &MockEmailSender{
				sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
					return &domain.EmailDeliveryResult{
						To:     req.To,
						Status: domain.StatusDelivered,
					}, nil
				},
			}

			useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

			result, err := useCase.SendWeatherUpdateEmail(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
		})
	}
}

func TestSendEmailUseCase_SendWeatherUpdateEmail_TemplateBuilderError(t *testing.T) {
	logger := logger.NewLogrusLogger()

	request := domain.WeatherUpdateEmailRequest{
		To:          "test@example.com",
		Subject:     "Weather Update",
		Name:        "User",
		City:        "Kyiv",
		Temperature: 15,
		Description: "Partly cloudy",
		Humidity:    65,
		WindSpeed:   12,
	}

	templateBuilder := &MockTemplateBuilder{
		buildWeatherUpdateEmailFunc: func(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error) {
			return "", errors.New("template builder error")
		},
	}

	emailSender := &MockEmailSender{
		sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "template builder error")
}

func TestSendEmailUseCase_SendWeatherUpdateEmail_EmailSenderError(t *testing.T) {
	logger := logger.NewLogrusLogger()

	request := domain.WeatherUpdateEmailRequest{
		To:          "test@example.com",
		Subject:     "Weather Update",
		Name:        "User",
		City:        "Kyiv",
		Temperature: 15,
		Description: "Partly cloudy",
		Humidity:    65,
		WindSpeed:   12,
	}

	templateBuilder := &MockTemplateBuilder{
		buildWeatherUpdateEmailFunc: func(ctx context.Context, email, city, description string, humidity int, windSpeed int, temperature int) (string, error) {
			return "<html><body>Test weather template</body></html>", nil
		},
	}

	emailSender := &MockEmailSender{
		sendEmailFunc: func(ctx context.Context, req domain.EmailRequest) (*domain.EmailDeliveryResult, error) {
			return nil, errors.New("email sender error")
		},
	}

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email sender error")
}

func TestSendEmailUseCase_Validation(t *testing.T) {
	logger := logger.NewLogrusLogger()
	templateBuilder := &MockTemplateBuilder{}
	emailSender := &MockEmailSender{}
	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	t.Run("Empty email in confirmation request", func(t *testing.T) {
		request := domain.ConfirmationEmailRequest{
			To:               "",
			Subject:          "Test",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost/confirm/token",
		}

		result, err := useCase.SendConfirmationEmail(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Empty city in weather request", func(t *testing.T) {
		request := domain.WeatherUpdateEmailRequest{
			To:          "test@example.com",
			Subject:     "Test",
			Name:        "User",
			City:        "",
			Temperature: 15,
			Description: "Clear",
			Humidity:    50,
			WindSpeed:   10,
		}

		result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
