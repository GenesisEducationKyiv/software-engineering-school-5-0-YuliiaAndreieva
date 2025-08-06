package unit

import (
	"context"
	"errors"
	"testing"

	"email/internal/core/domain"
	"email/internal/core/usecase"
	"email/tests/mocks"
	"shared/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			templateBuilder := &mocks.EmailTemplateBuilder{}
			emailSender := &mocks.EmailSender{}

			templateBuilder.On("BuildConfirmationEmail", mock.Anything, tt.request.To, tt.request.City, tt.request.ConfirmationLink).
				Return("<html><body>Test template</body></html>", nil)

			emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
				Return(&domain.EmailDeliveryResult{
					To:     tt.request.To,
					Status: domain.StatusDelivered,
				}, nil)

			useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

			result, err := useCase.SendConfirmationEmail(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.request.To, result.To)
			assert.Equal(t, tt.expectedStatus, result.Status)

			templateBuilder.AssertExpectations(t)
			emailSender.AssertExpectations(t)
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

	templateBuilder := &mocks.EmailTemplateBuilder{}
	templateBuilder.On("BuildConfirmationEmail", mock.Anything, request.To, request.City, request.ConfirmationLink).
		Return("", errors.New("template builder error"))

	emailSender := &mocks.EmailSender{}
	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendConfirmationEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "template builder error")

	templateBuilder.AssertExpectations(t)
}

func TestSendEmailUseCase_SendConfirmationEmail_EmailSenderError(t *testing.T) {
	logger := logger.NewLogrusLogger()

	request := domain.ConfirmationEmailRequest{
		To:               "test@example.com",
		Subject:          "Confirm Subscription",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token123",
	}

	templateBuilder := &mocks.EmailTemplateBuilder{}
	templateBuilder.On("BuildConfirmationEmail", mock.Anything, request.To, request.City, request.ConfirmationLink).
		Return("<html><body>Test template</body></html>", nil)

	emailSender := &mocks.EmailSender{}
	emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
		Return(nil, errors.New("email sender error"))

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendConfirmationEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email sender error")

	templateBuilder.AssertExpectations(t)
	emailSender.AssertExpectations(t)
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
			templateBuilder := &mocks.EmailTemplateBuilder{}
			emailSender := &mocks.EmailSender{}

			templateBuilder.On("BuildWeatherUpdateEmail", mock.Anything, tt.request.To, tt.request.City, tt.request.Description, tt.request.Humidity, tt.request.WindSpeed, tt.request.Temperature, mock.Anything).
				Return("<html><body>Test weather template</body></html>", nil)

			emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
				Return(&domain.EmailDeliveryResult{
					To:     tt.request.To,
					Status: domain.StatusDelivered,
				}, nil)

			useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

			result, err := useCase.SendWeatherUpdateEmail(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.request.To, result.To)
			assert.Equal(t, tt.expectedStatus, result.Status)

			templateBuilder.AssertExpectations(t)
			emailSender.AssertExpectations(t)
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

	templateBuilder := &mocks.EmailTemplateBuilder{}
	templateBuilder.On("BuildWeatherUpdateEmail", mock.Anything, request.To, request.City, request.Description, request.Humidity, request.WindSpeed, request.Temperature, mock.Anything).
		Return("", errors.New("template builder error"))

	emailSender := &mocks.EmailSender{}
	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "template builder error")

	templateBuilder.AssertExpectations(t)
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

	templateBuilder := &mocks.EmailTemplateBuilder{}
	templateBuilder.On("BuildWeatherUpdateEmail", mock.Anything, request.To, request.City, request.Description, request.Humidity, request.WindSpeed, request.Temperature, mock.Anything).
		Return("<html><body>Test weather template</body></html>", nil)

	emailSender := &mocks.EmailSender{}
	emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
		Return(nil, errors.New("email sender error"))

	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email sender error")

	templateBuilder.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

func TestSendEmailUseCase_Validation(t *testing.T) {
	logger := logger.NewLogrusLogger()
	templateBuilder := &mocks.EmailTemplateBuilder{}
	emailSender := &mocks.EmailSender{}
	useCase := usecase.NewSendEmailUseCase(emailSender, templateBuilder, logger, "http://localhost:8081")

	t.Run("Empty email in confirmation request", func(t *testing.T) {
		request := domain.ConfirmationEmailRequest{
			To:               "",
			Subject:          "Test",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost/confirm/token",
		}

		templateBuilder.On("BuildConfirmationEmail", mock.Anything, request.To, request.City, request.ConfirmationLink).
			Return("<html><body>Test template</body></html>", nil)

		emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
			Return(&domain.EmailDeliveryResult{
				To:     request.To,
				Status: domain.StatusDelivered,
			}, nil)

		result, err := useCase.SendConfirmationEmail(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		templateBuilder.AssertExpectations(t)
		emailSender.AssertExpectations(t)
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

		templateBuilder.On("BuildWeatherUpdateEmail", mock.Anything, request.To, request.City, request.Description, request.Humidity, request.WindSpeed, request.Temperature, mock.Anything).
			Return("<html><body>Test weather template</body></html>", nil)

		emailSender.On("SendEmail", mock.Anything, mock.AnythingOfType("domain.EmailRequest")).
			Return(&domain.EmailDeliveryResult{
				To:     request.To,
				Status: domain.StatusDelivered,
			}, nil)

		result, err := useCase.SendWeatherUpdateEmail(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		templateBuilder.AssertExpectations(t)
		emailSender.AssertExpectations(t)
	})
}
