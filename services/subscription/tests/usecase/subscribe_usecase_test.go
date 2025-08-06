package usecase

import (
	"context"
	"testing"

	"subscription/internal/config"
	"subscription/internal/core/domain"
	"subscription/internal/core/usecase"
	"subscription/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type subscribeUseCaseTestSetup struct {
	useCase            *usecase.SubscribeUseCase
	mockRepo           *mocks.SubscriptionRepository
	mockTokenService   *mocks.TokenService
	mockEventPublisher *mocks.EventPublisher
	mockLogger         *mocks.Logger
	config             *config.Config
}

func setupSubscribeUseCaseTest() *subscribeUseCaseTestSetup {
	mockRepo := &mocks.SubscriptionRepository{}
	mockTokenService := &mocks.TokenService{}
	mockEventPublisher := &mocks.EventPublisher{}
	mockLogger := &mocks.Logger{}

	config := &config.Config{
		Token: config.TokenConfig{
			Expiration: "24h",
		},
		Server: config.ServerConfig{
			BaseURL: "http://subscription-service:8082",
		},
	}

	useCase := usecase.NewSubscribeUseCase(mockRepo, mockTokenService, mockEventPublisher, mockLogger, config)
	typedUseCase, ok := useCase.(*usecase.SubscribeUseCase)
	if !ok {
		panic("Failed to type assert SubscribeUseCase")
	}

	return &subscribeUseCaseTestSetup{
		useCase:            typedUseCase,
		mockRepo:           mockRepo,
		mockTokenService:   mockTokenService,
		mockEventPublisher: mockEventPublisher,
		mockLogger:         mockLogger,
		config:             config,
	}
}

func (ts *subscribeUseCaseTestSetup) setupSuccessMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockEventPublisher.On("PublishSubscriptionCreated", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *subscribeUseCaseTestSetup) setupDuplicateMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(domain.ErrDuplicateSubscription)
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func (ts *subscribeUseCaseTestSetup) setupRepositoryErrorMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(assert.AnError)
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func (ts *subscribeUseCaseTestSetup) setupTokenErrorMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("", assert.AnError)
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *subscribeUseCaseTestSetup) setupCreateErrorMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(assert.AnError)
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
}

func TestSubscribeUseCase_Success(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	t.Run("Valid subscription request", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupSuccessMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "test-token", result.Token)
		assert.Contains(t, result.Message, "Subscription successful")

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockEventPublisher.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscribeUseCase_DuplicateSubscription(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	t.Run("Duplicate subscription", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupDuplicateMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Subscription already exists")

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscribeUseCase_RepositoryError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	t.Run("Repository error", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupRepositoryErrorMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscribeUseCase_TokenGenerationError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	t.Run("Token generation error", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupTokenErrorMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		ts.mockTokenService.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscribeUseCase_CreateSubscriptionError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	t.Run("Create subscription error", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupCreateErrorMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}
