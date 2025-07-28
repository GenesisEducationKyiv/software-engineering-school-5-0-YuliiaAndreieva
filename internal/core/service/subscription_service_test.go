//go:build unit
// +build unit

package service

import (
	"context"
	"testing"
	"weather-api/internal/core/domain"
	"weather-api/internal/mocks"

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

func TestSubscriptionServiceImpl_CreateSubscription_Success(t *testing.T) {
	// Arrange
	mockTokenSvc := &mocks.MockTokenService{}
	mockRepo := &mocks.MockSubscriptionRepository{}
	mockCityRepo := &mocks.MockCityRepo{}
	mockWeatherProvider := &mocks.MockWeatherProvider{}
	mockEmailService := &MockEmailNotifier{}

	service := NewSubscriptionService(
		mockRepo,
		mockCityRepo,
		mockWeatherProvider,
		mockTokenSvc,
		mockEmailService,
	)

	expectedToken := "test-token-123"
	email := "test@example.com"
	cityID := int64(1)
	frequency := domain.FrequencyDaily

	mockTokenSvc.On("GenerateToken").Return(expectedToken, nil)
	mockRepo.On("CreateSubscription", mock.Anything, domain.Subscription{
		Email:       email,
		CityID:      cityID,
		Frequency:   frequency,
		Token:       expectedToken,
		IsConfirmed: false,
	}).Return(nil)

	// Act
	token, err := service.CreateSubscription(context.Background(), email, cityID, frequency)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)

	mockTokenSvc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
