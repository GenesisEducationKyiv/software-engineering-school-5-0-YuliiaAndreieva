//go:build unit
// +build unit

package service

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/core/domain"
	"weather-api/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionServiceImpl_CreateSubscription_TokenGenerationFailed(t *testing.T) {
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

	email := "test@example.com"
	cityID := int64(1)
	frequency := domain.FrequencyDaily

	mockTokenSvc.On("GenerateToken").Return("", errors.New("token generation failed"))

	// Act
	token, err := service.CreateSubscription(context.Background(), email, cityID, frequency)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "unable to generate token: token generation failed", err.Error())
	assert.Equal(t, "", token)

	mockTokenSvc.AssertExpectations(t)
}

func TestSubscriptionServiceImpl_CreateSubscription_RepositoryCreationFailed(t *testing.T) {
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
	}).Return(errors.New("database error"))

	// Act
	token, err := service.CreateSubscription(context.Background(), email, cityID, frequency)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "unable to create subscription in repository: database error", err.Error())
	assert.Equal(t, "", token)

	mockTokenSvc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
