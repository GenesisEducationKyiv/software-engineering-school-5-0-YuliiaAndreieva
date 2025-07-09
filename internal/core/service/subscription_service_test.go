//go:build unit
// +build unit

package service

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/mocks"

	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailNotifier struct {
	mock.Mock
}

func (m *MockEmailNotifier) SendConfirmationEmail(subscription *domain.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockEmailNotifier) SendUpdates(updates []domain.WeatherUpdate) error {
	args := m.Called(updates)
	return args.Error(0)
}

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
			emailNotifier *MockEmailNotifier,
			tokenSvc *mocks.MockTokenService,
		)
		expectedToken string
		expectedErr   error
	}{
		{
			name: "happy path – city already in database",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailNotifier *MockEmailNotifier,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).Return(cityRow, nil)

				subRepo.On("IsSubscriptionExists", ctx, ports.IsSubscriptionExistsOptions{
					Email:     email,
					CityID:    cityRow.ID,
					Frequency: domain.FrequencyDaily,
				}).Return(false, nil)

				tokenSvc.On("GenerateToken").Return(tokenStr, nil)

				sub := domain.Subscription{
					Email:       email,
					CityID:      cityRow.ID,
					Frequency:   domain.FrequencyDaily,
					Token:       tokenStr,
					IsConfirmed: false,
				}
				subRepo.On("CreateSubscription", ctx, sub).Return(nil)

				expectedSub := sub
				expectedSub.City = &cityRow
				emailNotifier.On("SendConfirmationEmail", &expectedSub).Return(nil)
			},
			expectedToken: tokenStr,
			expectedErr:   nil,
		},
		{
			name: "city is not in database – CheckCityExists returns nil",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailNotifier *MockEmailNotifier,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).
					Return(domain.City{}, domain.ErrCityNotFound)

				weatherProv.On("CheckCityExists", ctx, cityName).Return(nil)

				cityRepo.On("Create", ctx, domain.City{Name: cityName}).
					Return(cityRow, nil)

				subRepo.On("IsSubscriptionExists", ctx, ports.IsSubscriptionExistsOptions{
					Email:     email,
					CityID:    cityRow.ID,
					Frequency: domain.FrequencyDaily,
				}).Return(false, nil)

				tokenSvc.On("GenerateToken").Return(tokenStr, nil)

				sub := domain.Subscription{
					Email:       email,
					CityID:      cityRow.ID,
					Frequency:   domain.FrequencyDaily,
					Token:       tokenStr,
					IsConfirmed: false,
				}
				subRepo.On("CreateSubscription", ctx, sub).Return(nil)

				expectedSub := sub
				expectedSub.City = &cityRow
				emailNotifier.On("SendConfirmationEmail", &expectedSub).Return(nil)
			},
			expectedToken: tokenStr,
			expectedErr:   nil,
		},
		{
			name: "email already subscribed",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailNotifier *MockEmailNotifier,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).Return(cityRow, nil)
				subRepo.On("IsSubscriptionExists", ctx, ports.IsSubscriptionExistsOptions{
					Email:     email,
					CityID:    cityRow.ID,
					Frequency: domain.FrequencyDaily,
				}).Return(true, nil)
			},
			expectedToken: "",
			expectedErr:   domain.ErrEmailAlreadySubscribed,
		},
		{
			name: "ValidateCity returns not found",
			setupMocks: func(subRepo *mocks.MockSubscriptionRepository, cityRepo *mocks.MockCityRepo,
				weatherProv *mocks.MockWeatherProvider, emailNotifier *MockEmailNotifier,
				tokenSvc *mocks.MockTokenService) {

				cityRepo.On("GetByName", ctx, cityName).
					Return(domain.City{}, domain.ErrCityNotFound)
				weatherProv.On("CheckCityExists", ctx, cityName).
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
			emailNotifier := &MockEmailNotifier{}
			tokenSvc := &mocks.MockTokenService{}

			tt.setupMocks(subRepo, cityRepo, weatherProv, emailNotifier, tokenSvc)

			s := NewSubscriptionService(
				subRepo, cityRepo, weatherProv, tokenSvc, emailNotifier,
			)

			token, err := s.Subscribe(ctx, ports.SubscribeOptions{
				Email:     email,
				City:      cityName,
				Frequency: domain.FrequencyDaily,
			})
			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedErr, err)

			subRepo.AssertExpectations(t)
			cityRepo.AssertExpectations(t)
			weatherProv.AssertExpectations(t)
			emailNotifier.AssertExpectations(t)
			tokenSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_Confirm(t *testing.T) {
	ctx := context.Background()
	const token = "tok123"

	tests := []struct {
		name       string
		setupMocks func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService)
		expectErr  error
	}{
		{
			name: "successfully confirms subscription",
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(nil)
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
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(domain.ErrTokenNotFound)
			},
			expectErr: domain.ErrTokenNotFound,
		},
		{
			name: "update fails",
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(nil)
				sub := domain.Subscription{Token: token, IsConfirmed: false}
				r.On("GetSubscriptionByToken", ctx, token).Return(sub, nil)
				subConfirmed := sub
				subConfirmed.IsConfirmed = true
				r.On("UpdateSubscription", ctx, subConfirmed).
					Return(errors.New("db error"))
			},
			expectErr: errors.New("unable to update subscription confirmation: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			cityRepo := &mocks.MockCityRepo{}
			weatherProv := &mocks.MockWeatherProvider{}
			emailNotifier := &MockEmailNotifier{}
			tokenSvc := &mocks.MockTokenService{}

			tt.setupMocks(repo, tokenSvc)

			s := NewSubscriptionService(
				repo, cityRepo, weatherProv, tokenSvc, emailNotifier,
			)

			err := s.Confirm(ctx, token)
			assert.Equal(t, tt.expectErr, err)
			repo.AssertExpectations(t)
			tokenSvc.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	const token = "tok123"

	tests := []struct {
		name       string
		setupMocks func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService)
		expectErr  error
	}{
		{
			name: "successfully unsubscribes",
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(nil)
				r.On("DeleteSubscription", ctx, token).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "token not found",
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(domain.ErrTokenNotFound)
			},
			expectErr: domain.ErrTokenNotFound,
		},
		{
			name: "delete fails",
			setupMocks: func(r *mocks.MockSubscriptionRepository, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("CheckTokenExists", ctx, token).Return(nil)
				r.On("DeleteSubscription", ctx, token).
					Return(errors.New("db error"))
			},
			expectErr: errors.New("unable to delete subscription: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			cityRepo := &mocks.MockCityRepo{}
			weatherProv := &mocks.MockWeatherProvider{}
			emailNotifier := &MockEmailNotifier{}
			tokenSvc := &mocks.MockTokenService{}

			tt.setupMocks(repo, tokenSvc)

			s := NewSubscriptionService(
				repo, cityRepo, weatherProv, tokenSvc, emailNotifier,
			)

			err := s.Unsubscribe(ctx, token)
			assert.Equal(t, tt.expectErr, err)
			repo.AssertExpectations(t)
			tokenSvc.AssertExpectations(t)
		})
	}
}
