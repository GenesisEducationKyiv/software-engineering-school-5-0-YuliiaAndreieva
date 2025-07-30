package tests

import (
	"context"
	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"

	"github.com/stretchr/testify/mock"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateSubscription(ctx context.Context, subscription domain.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetSubscriptionByToken(ctx context.Context, token string) (*domain.Subscription, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UpdateSubscription(ctx context.Context, subscription domain.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteSubscription(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) IsSubscriptionExists(ctx context.Context, email, city string) (bool, error) {
	args := m.Called(ctx, email, city)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) ListByFrequency(ctx context.Context, frequency string, lastID, pageSize int) ([]domain.Subscription, error) {
	args := m.Called(ctx, frequency, lastID, pageSize)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(ctx context.Context, email, duration string) (string, error) {
	args := m.Called(ctx, email, duration)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateToken(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (out.EmailDeliveryResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(out.EmailDeliveryResult), args.Error(1)
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
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func SetupSuccessLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
}

func SetupWarningLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return()
}

func SetupErrorLoggerMocks(logger *MockLogger) {
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func SetupJSONErrorLoggerMocks(logger *MockLogger) {
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
}
