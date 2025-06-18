package service_test

import (
	"testing"
	"weather-api/internal/util/emailutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"weather-api/internal/core/domain"
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
				subKyiv, bodyKyiv := emailutil.BuildWeatherUpdateEmail("Kyiv", 20.5, 60, "Sunny", "token1")
				subLviv, bodyLviv := emailutil.BuildWeatherUpdateEmail("Lviv", 18.0, 65, "Cloudy", "token2")

				es.On("SendEmail", "user1@example.com", subKyiv, bodyKyiv).Return(nil).Once()
				es.On("SendEmail", "user2@example.com", subLviv, bodyLviv).Return(nil).Once()
			},
			verifyMocks: func(t *testing.T, es *mocks.MockEmailService) {
				es.AssertExpectations(t)
			},
		},
		{
			name: "smtp error is ignored",
			updates: []domain.WeatherUpdate{
				{
					Subscription: domain.Subscription{
						Email: "user1@example.com",
						City:  &domain.City{Name: "Kyiv"},
						Token: "token1",
					},
					Weather: domain.Weather{Temperature: 20.5, Humidity: 60, Description: "Sunny"},
				},
			},
			setupMocks: func(es *mocks.MockEmailService) {
				subj, body := emailutil.BuildWeatherUpdateEmail("Kyiv", 20.5, 60, "Sunny", "token1")
				es.On("SendEmail", "user1@example.com", subj, body).
					Return(assert.AnError).Once()
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
				es.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailMock := &mocks.MockEmailService{}
			tt.setupMocks(emailMock)

			repoMock := &mocks.MockSubscriptionRepository{}
			wsMock := &mocks.MockWeatherProvider{}

			s := service.NewEmailService(repoMock, wsMock, emailMock)

			err := s.SendUpdates(tt.updates)
			assert.NoError(t, err)

			tt.verifyMocks(t, emailMock)
		})
	}
}
