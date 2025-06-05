package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"weather-api/internal/core/domain"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) GetSubscriptionsByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error) {
	args := m.Called(ctx, frequency)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CreateSubscription(ctx context.Context, sub domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetSubscriptionByToken(ctx context.Context, token string) (domain.Subscription, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UpdateSubscription(ctx context.Context, sub domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteSubscription(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) IsEmailSubscribed(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) IsTokenExists(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

type MockWeatherService struct {
	mock.Mock
}

func (m *MockWeatherService) GetWeather(city string) (domain.Weather, error) {
	args := m.Called(city)
	return args.Get(0).(domain.Weather), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
