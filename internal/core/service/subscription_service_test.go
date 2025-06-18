package service

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/mocks"
	"weather-api/internal/util/emailutil"

	"weather-api/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionService_Subscribe(t *testing.T) {
	ctx := context.Background()
	const (
		email    = "user@example.com"
		cityName = "Kyiv"
		tokenStr = "token123"
	)
	cityRow := domain.City{ID: 1, Name: cityName}

	tests := []struct {
		name       string
		setupMocks func(
			subRepo *mocks.MockSubscriptionRepository,
			cityRepo *mocks.MockCityRepo,
			weatherProv *mocks.MockWeatherProvider,
			emailSvc *mocks.MockEmailService,
			tokenSvc *mocks.MockTokenService,
		)
		expectedToken string
		expectedErr   error
	}{
		{
			name: "happy path – city already cached",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailSvc *mocks.MockEmailService,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).Return(cityRow, nil)

				subRepo.On("IsSubscriptionExists", ctx, email, cityRow.ID, domain.FrequencyDaily).
					Return(false, nil)

				tokenSvc.On("GenerateToken").Return(tokenStr, nil)

				sub := domain.Subscription{
					Email:       email,
					CityID:      cityRow.ID,
					Frequency:   domain.FrequencyDaily,
					Token:       tokenStr,
					IsConfirmed: false,
				}
				subRepo.On("CreateSubscription", ctx, sub).Return(nil)

				subject, body := emailutil.BuildConfirmationEmail(cityName, tokenStr)
				emailSvc.On("SendEmail", email, subject, body).Return(nil)
			},
			expectedToken: tokenStr,
			expectedErr:   nil,
		},
		{
			name: "city not cached – ValidateCity succeeds",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailSvc *mocks.MockEmailService,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).
					Return(domain.City{}, domain.ErrCityNotFound)

				weatherProv.On("ValidateCity", ctx, cityName).Return(nil)

				cityRepo.On("Create", ctx, domain.City{Name: cityName}).
					Return(cityRow, nil)

				subRepo.On("IsSubscriptionExists", ctx, email, cityRow.ID, domain.FrequencyDaily).
					Return(false, nil)

				tokenSvc.On("GenerateToken").Return(tokenStr, nil)

				sub := domain.Subscription{
					Email:       email,
					CityID:      cityRow.ID,
					Frequency:   domain.FrequencyDaily,
					Token:       tokenStr,
					IsConfirmed: false,
				}
				subRepo.On("CreateSubscription", ctx, sub).Return(nil)

				subject, body := emailutil.BuildConfirmationEmail(cityName, tokenStr)
				emailSvc.On("SendEmail", email, subject, body).Return(nil)
			},
			expectedToken: tokenStr,
			expectedErr:   nil,
		},
		{
			name: "email already subscribed",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailSvc *mocks.MockEmailService,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).Return(cityRow, nil)
				subRepo.On("IsSubscriptionExists", ctx, email, cityRow.ID, domain.FrequencyDaily).
					Return(true, nil)
			},
			expectedToken: "",
			expectedErr:   domain.ErrEmailAlreadySubscribed,
		},
		{
			name: "ValidateCity returns not found",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailSvc *mocks.MockEmailService,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).
					Return(domain.City{}, domain.ErrCityNotFound)
				weatherProv.On("ValidateCity", ctx, cityName).
					Return(domain.ErrCityNotFound)
			},
			expectedToken: "",
			expectedErr:   domain.ErrCityNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := &mocks.MockSubscriptionRepository{}
			cityRepo := &mocks.MockCityRepo{}
			weatherProv := &mocks.MockWeatherProvider{}
			emailSvc := &mocks.MockEmailService{}
			tokenSvc := &mocks.MockTokenService{}
			weatherSvc := &mocks.MockWeatherProvider{}

			tt.setupMocks(subRepo, cityRepo, weatherProv, emailSvc, tokenSvc)

			s := NewSubscriptionService(
				subRepo, cityRepo,
				weatherSvc,
				weatherProv,
				emailSvc, tokenSvc,
			)

			token, err := s.Subscribe(ctx, email, cityName, domain.FrequencyDaily)
			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedErr, err)

			subRepo.AssertExpectations(t)
			cityRepo.AssertExpectations(t)
			weatherProv.AssertExpectations(t)
			emailSvc.AssertExpectations(t)
			tokenSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_Confirm(t *testing.T) {
	ctx := context.Background()
	const token = "tok123"

	tests := []struct {
		name       string
		setupMocks func(r *mocks.MockSubscriptionRepository)
		expectErr  error
	}{
		{
			name: "successfully confirms subscription",
			setupMocks: func(r *mocks.MockSubscriptionRepository) {
				r.On("IsTokenExists", ctx, token).Return(true, nil)
				sub := domain.Subscription{Token: token, IsConfirmed: false}
				r.On("GetSubscriptionByToken", ctx, token).Return(sub, nil)
				subConfirmed := sub
				subConfirmed.IsConfirmed = true
				r.On("UpdateSubscription", ctx, subConfirmed).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "token not found",
			setupMocks: func(r *mocks.MockSubscriptionRepository) {
				r.On("IsTokenExists", ctx, token).Return(false, nil)
			},
			expectErr: domain.ErrTokenNotFound,
		},
		{
			name: "update fails",
			setupMocks: func(r *mocks.MockSubscriptionRepository) {
				r.On("IsTokenExists", ctx, token).Return(true, nil)
				sub := domain.Subscription{Token: token, IsConfirmed: false}
				r.On("GetSubscriptionByToken", ctx, token).Return(sub, nil)
				subConfirmed := sub
				subConfirmed.IsConfirmed = true
				r.On("UpdateSubscription", ctx, subConfirmed).
					Return(errors.New("db error"))
			},
			expectErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			cityRepo := &mocks.MockCityRepo{}
			wp := &mocks.MockWeatherProvider{}
			emailSvc := &mocks.MockEmailService{}
			tokenSvc := &mocks.MockTokenService{}
			weatherSvc := &mocks.MockWeatherService{}

			tt.setupMocks(repo)

			s := NewSubscriptionService(
				repo, cityRepo, weatherSvc, wp, emailSvc, tokenSvc,
			)

			err := s.Confirm(ctx, token)
			assert.Equal(t, tt.expectErr, err)
			repo.AssertExpectations(t)
		})
	}
}
