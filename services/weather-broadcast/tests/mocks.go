package tests

import (
	"context"
	"weather-broadcast-service/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockSubscriptionClient struct {
	mock.Mock
}

func (m *MockSubscriptionClient) ListByFrequency(ctx context.Context, query domain.ListSubscriptionsQuery) (*domain.SubscriptionList, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*domain.SubscriptionList), args.Error(1)
}

type MockWeatherClient struct {
	mock.Mock
}

func (m *MockWeatherClient) GetWeatherByCity(ctx context.Context, city string) (*domain.Weather, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(*domain.Weather), args.Error(1)
}

type MockWeatherMailer struct {
	mock.Mock
}

func (m *MockWeatherMailer) SendWeather(ctx context.Context, info *domain.WeatherMailSuccessInfo) error {
	args := m.Called(ctx, info)
	return args.Error(0)
}

func (m *MockWeatherMailer) SendError(ctx context.Context, info *domain.WeatherMailErrorInfo) error {
	args := m.Called(ctx, info)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func SetupCommonLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
}

func SetupSuccessLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
}

func SetupErrorLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func SetupWarningLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
}
