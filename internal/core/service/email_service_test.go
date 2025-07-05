//go:build unit
// +build unit

package service_test

import (
	"errors"
	"testing"
	"weather-api/internal/util/emailutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"
	"weather-api/internal/core/service"
	"weather-api/internal/mocks"
)

func TestEmailService_SendUpdates(t *testing.T) {
	tests := []struct {
		name        string
		updates     []domain.WeatherUpdate
		setupMocks  func(emailSvc *mocks.MockEmailService)
		verifyMocks func(t *testing.T, emailSvc *mocks.MockEmailService)
	}{
		{
			name: "all emails sent successfully",
			updates: []domain.WeatherUpdate{
				{
					Subscription: domain.Subscription{
						Email: "user1@example.com",
						City:  &domain.City{Name: "Kyiv"},
						Token: "token1",
					},
					Weather: domain.Weather{Temperature: 20.5, Humidity: 60, Description: "Sunny"},
				},
				{
					Subscription: domain.Subscription{
						Email: "user2@example.com",
						City:  &domain.City{Name: "Lviv"},
						Token: "token2",
					},
					Weather: domain.Weather{Temperature: 18.0, Humidity: 65, Description: "Cloudy"},
				},
			},
			setupMocks: func(es *mocks.MockEmailService) {
				subKyiv, bodyKyiv := emailutil.BuildWeatherUpdateEmail(emailutil.WeatherUpdateEmailOptions{
					City:        "Kyiv",
					Temperature: 20.5,
					Humidity:    60,
					Description: "Sunny",
					Token:       "token1",
				})
				subLviv, bodyLviv := emailutil.BuildWeatherUpdateEmail(emailutil.WeatherUpdateEmailOptions{
					City:        "Lviv",
					Temperature: 18.0,
					Humidity:    65,
					Description: "Cloudy",
					Token:       "token2",
				})

				es.On("SendEmail", ports.SendEmailOptions{
					To:      "user1@example.com",
					Subject: subKyiv,
					Body:    bodyKyiv,
				}).Return(nil).Once()
				es.On("SendEmail", ports.SendEmailOptions{
					To:      "user2@example.com",
					Subject: subLviv,
					Body:    bodyLviv,
				}).Return(nil).Once()
			},
			verifyMocks: func(t *testing.T, es *mocks.MockEmailService) {
				es.AssertExpectations(t)
			},
		},
		{
			name:       "no updates",
			updates:    []domain.WeatherUpdate{},
			setupMocks: func(*mocks.MockEmailService) {},
			verifyMocks: func(t *testing.T, es *mocks.MockEmailService) {
				es.AssertNotCalled(t, "SendEmail", mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailMock := &mocks.MockEmailService{}
			tt.setupMocks(emailMock)

			s := service.NewEmailService(emailMock)

			err := s.SendUpdates(tt.updates)
			assert.NoError(t, err)

			tt.verifyMocks(t, emailMock)
		})
	}
}

func TestEmailService_SendConfirmationEmail(t *testing.T) {
	tests := []struct {
		name         string
		subscription *domain.Subscription
		setupMocks   func(emailSvc *mocks.MockEmailService)
		expectErr    error
	}{
		{
			name: "successfully sends confirmation email",
			subscription: &domain.Subscription{
				Email: "user@example.com",
				City:  &domain.City{Name: "Kyiv"},
				Token: "token123",
			},
			setupMocks: func(es *mocks.MockEmailService) {
				subject, body := emailutil.BuildConfirmationEmail("Kyiv", "token123")
				es.On("SendEmail", ports.SendEmailOptions{
					To:      "user@example.com",
					Subject: subject,
					Body:    body,
				}).Return(nil).Once()
			},
			expectErr: nil,
		},
		{
			name: "email service returns error",
			subscription: &domain.Subscription{
				Email: "user@example.com",
				City:  &domain.City{Name: "Kyiv"},
				Token: "token123",
			},
			setupMocks: func(es *mocks.MockEmailService) {
				subject, body := emailutil.BuildConfirmationEmail("Kyiv", "token123")
				es.On("SendEmail", ports.SendEmailOptions{
					To:      "user@example.com",
					Subject: subject,
					Body:    body,
				}).Return(assert.AnError).Once()
			},
			expectErr: errors.New("unable to send confirmation email to user@example.com: assert.AnError general error for testing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailMock := &mocks.MockEmailService{}
			tt.setupMocks(emailMock)

			s := service.NewEmailService(emailMock)

			err := s.SendConfirmationEmail(tt.subscription)
			assert.Equal(t, tt.expectErr, err)

			emailMock.AssertExpectations(t)
		})
	}
}
