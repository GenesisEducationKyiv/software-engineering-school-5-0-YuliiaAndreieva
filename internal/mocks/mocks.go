package mocks

import (
	"context"
	"weather-api/internal/core/domain"

	"github.com/stretchr/testify/mock"
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

func (m *MockSubscriptionRepository) UpdateLastSentAt(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
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

func (m *MockSubscriptionRepository) IsSubscriptionExists(ctx context.Context, email string, cityID int64, frequency domain.Frequency) (bool, error) {
	args := m.Called(ctx, email, cityID, frequency)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) IsTokenExists(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

type MockWeatherProvider struct{ mock.Mock }

func (m *MockWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(domain.Weather), args.Error(1)
}
func (m *MockWeatherProvider) ValidateCity(ctx context.Context, city string) error {
	args := m.Called(ctx, city)
	return args.Error(0)
}

type MockWeatherService struct{ mock.Mock }

func (s *MockWeatherService) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	args := s.Called(ctx, city)
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

type MockCityRepo struct{ mock.Mock }

func (m *MockCityRepo) GetByName(ctx context.Context, name string) (domain.City, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(domain.City), args.Error(1)
}
func (m *MockCityRepo) Create(ctx context.Context, city domain.City) (domain.City, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(domain.City), args.Error(1)
}
