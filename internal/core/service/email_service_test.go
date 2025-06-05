package service

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/mocks"

	"github.com/stretchr/testify/mock"

	"weather-api/internal/core/domain"
	"weather-api/internal/util"
)

func TestEmailService_SendUpdates(t *testing.T) {
	ctx := context.Background()
	frequency := domain.FrequencyDaily

	tests := []struct {
		name           string
		setupMocks     func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService)
		verifyMocks    func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService)
		subscriptions  []domain.Subscription
		repoError      error
		weatherResults map[string]struct {
			weather domain.Weather
			err     error
		}
		emailResults map[string]struct {
			err error
		}
	}{
		{
			name: "success with confirmed subscriptions",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				subs := []domain.Subscription{
					{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: true},
					{Email: "user2@example.com", City: "Lviv", Frequency: frequency, Token: "token2", IsConfirmed: true},
					{Email: "user3@example.com", City: "Odesa", Frequency: frequency, Token: "token3", IsConfirmed: false},
				}
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return(subs, nil)
				weatherSvc.On("GetWeather", "Kyiv").Return(domain.Weather{Temperature: 20.5, Humidity: 60, Description: "Sunny"}, nil)
				weatherSvc.On("GetWeather", "Lviv").Return(domain.Weather{Temperature: 18.0, Humidity: 65, Description: "Cloudy"}, nil)
				subjectKyiv, bodyKyiv := util.BuildWeatherUpdateEmail("Kyiv", 20.5, 60, "Sunny", "token1")
				subjectLviv, bodyLviv := util.BuildWeatherUpdateEmail("Lviv", 18.0, 65, "Cloudy", "token2")
				emailSvc.On("SendEmail", "user1@example.com", subjectKyiv, bodyKyiv).Return(nil)
				emailSvc.On("SendEmail", "user2@example.com", subjectLviv, bodyLviv).Return(nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertExpectations(t)
				emailSvc.AssertExpectations(t)
				emailSvc.AssertNotCalled(t, "SendEmail", "user3@example.com", mock.Anything, mock.Anything)
			},
		},
		{
			name: "repository error",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return([]domain.Subscription(nil), errors.New("db error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "weather service error",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				subs := []domain.Subscription{
					{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: true},
				}
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return(subs, nil)
				weatherSvc.On("GetWeather", "Kyiv").Return(domain.Weather{}, errors.New("API error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertExpectations(t)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "email service error",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				subs := []domain.Subscription{
					{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: true},
				}
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return(subs, nil)
				weather := domain.Weather{Temperature: 20.5, Humidity: 60, Description: "Sunny"}
				weatherSvc.On("GetWeather", "Kyiv").Return(weather, nil)
				subject, body := util.BuildWeatherUpdateEmail("Kyiv", 20.5, 60, "Sunny", "token1")
				emailSvc.On("SendEmail", "user1@example.com", subject, body).Return(errors.New("SMTP error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertExpectations(t)
				emailSvc.AssertExpectations(t)
			},
		},
		{
			name: "empty subscriptions",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return([]domain.Subscription{}, nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "all subscriptions unconfirmed",
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				subs := []domain.Subscription{
					{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: false},
					{Email: "user2@example.com", City: "Lviv", Frequency: frequency, Token: "token2", IsConfirmed: false},
				}
				repo.On("GetSubscriptionsByFrequency", ctx, string(frequency)).Return(subs, nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			weatherSvc := &mocks.MockWeatherService{}
			emailSvc := &mocks.MockEmailService{}
			service := NewEmailService(repo, weatherSvc, emailSvc)

			tt.setupMocks(repo, weatherSvc, emailSvc)

			service.SendUpdates(ctx, frequency)

			tt.verifyMocks(t, repo, weatherSvc, emailSvc)
		})
	}
}

func TestEmailService_sendUpdates(t *testing.T) {
	frequency := domain.FrequencyDaily

	tests := []struct {
		name          string
		subscriptions []domain.Subscription
		setupMocks    func(weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService)
		verifyMocks   func(t *testing.T, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService)
	}{
		{
			name: "success with mixed subscriptions",
			subscriptions: []domain.Subscription{
				{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: true},
				{Email: "user2@example.com", City: "Lviv", Frequency: frequency, Token: "token2", IsConfirmed: false},
			},
			setupMocks: func(weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				weatherSvc.On("GetWeather", "Kyiv").Return(domain.Weather{Temperature: 20.5, Humidity: 60, Description: "Sunny"}, nil)
				subject, body := util.BuildWeatherUpdateEmail("Kyiv", 20.5, 60, "Sunny", "token1")
				emailSvc.On("SendEmail", "user1@example.com", subject, body).Return(nil)
			},
			verifyMocks: func(t *testing.T, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				weatherSvc.AssertExpectations(t)
				emailSvc.AssertExpectations(t)
				emailSvc.AssertNotCalled(t, "SendEmail", "user2@example.com", mock.Anything, mock.Anything)
			},
		},
		{
			name:          "empty subscriptions",
			subscriptions: []domain.Subscription{},
			setupMocks:    func(weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {},
			verifyMocks: func(t *testing.T, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "weather error stops processing",
			subscriptions: []domain.Subscription{
				{Email: "user1@example.com", City: "Kyiv", Frequency: frequency, Token: "token1", IsConfirmed: true},
				{Email: "user2@example.com", City: "Lviv", Frequency: frequency, Token: "token2", IsConfirmed: true},
			},
			setupMocks: func(weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				weatherSvc.On("GetWeather", "Kyiv").Return(domain.Weather{}, errors.New("API error"))
			},
			verifyMocks: func(t *testing.T, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService) {
				weatherSvc.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", "Lviv")
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			weatherSvc := &mocks.MockWeatherService{}
			emailSvc := &mocks.MockEmailService{}
			service := NewEmailService(repo, weatherSvc, emailSvc)

			tt.setupMocks(weatherSvc, emailSvc)

			service.sendUpdates(tt.subscriptions)

			tt.verifyMocks(t, weatherSvc, emailSvc)
		})
	}
}
