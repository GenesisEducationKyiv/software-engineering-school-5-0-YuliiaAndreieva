package usecase

import (
	"context"
	"testing"

	"subscription/internal/core/domain"
	"subscription/internal/core/ports/out"
	"subscription/internal/core/usecase"
	"subscription/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type subscribeUseCaseTestSetup struct {
	useCase          *usecase.SubscribeUseCase
	mockRepo         *tests.MockSubscriptionRepository
	mockTokenService *tests.MockTokenService
	mockEmailService *tests.MockEmailService
	mockLogger       *tests.MockLogger
}

func setupSubscribeUseCaseTest() *subscribeUseCaseTestSetup {
	mockRepo := &tests.MockSubscriptionRepository{}
	mockTokenService := &tests.MockTokenService{}
	mockEmailService := &tests.MockEmailService{}
	mockLogger := &tests.MockLogger{}

	useCase := usecase.NewSubscribeUseCase(mockRepo, mockTokenService, mockEmailService, mockLogger).(*usecase.SubscribeUseCase)

	return &subscribeUseCaseTestSetup{
		useCase:          useCase,
		mockRepo:         mockRepo,
		mockTokenService: mockTokenService,
		mockEmailService: mockEmailService,
		mockLogger:       mockLogger,
	}
}

func (ts *subscribeUseCaseTestSetup) setupSuccessMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("CreateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(out.EmailDeliveryResult{}, nil)
	tests.SetupSuccessLoggerMocks(ts.mockLogger)
}

func (ts *subscribeUseCaseTestSetup) setupDuplicateMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(true, nil)
	tests.SetupWarningLoggerMocks(ts.mockLogger)
}

func (ts *subscribeUseCaseTestSetup) setupRepositoryErrorMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, assert.AnError)
	tests.SetupErrorLoggerMocks(ts.mockLogger)
}

func (ts *subscribeUseCaseTestSetup) setupTokenErrorMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("", assert.AnError)
	tests.SetupErrorLoggerMocks(ts.mockLogger)
}

func (ts *subscribeUseCaseTestSetup) setupCreateErrorMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("CreateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(assert.AnError)
	tests.SetupErrorLoggerMocks(ts.mockLogger)
}

func (ts *subscribeUseCaseTestSetup) setupEmailErrorMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("CreateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(out.EmailDeliveryResult{}, assert.AnError)
	tests.SetupErrorLoggerMocks(ts.mockLogger)
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
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
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Subscription successful")

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockEmailService.AssertExpectations(t)
	})

	t.Run("Another valid subscription request", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "user@test.com",
			City:      "Lviv",
			Frequency: "weekly",
		}

		ts.setupSuccessMocks(request)

		result, err := ts.useCase.Subscribe(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Token)
		assert.Contains(t, result.Message, "Subscription successful")
	})
}

func TestSubscribeUseCase_DuplicateSubscription(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	request := domain.SubscriptionRequest{
		Email:     "duplicate@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupDuplicateMocks(request)

	result, err := ts.useCase.Subscribe(context.Background(), request)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "already subscribed")
}

func TestSubscribeUseCase_RepositoryError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	request := domain.SubscriptionRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupRepositoryErrorMocks(request)

	result, err := ts.useCase.Subscribe(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSubscribeUseCase_TokenGenerationError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	request := domain.SubscriptionRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupTokenErrorMocks(request)

	result, err := ts.useCase.Subscribe(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSubscribeUseCase_CreateSubscriptionError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	request := domain.SubscriptionRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupCreateErrorMocks(request)

	result, err := ts.useCase.Subscribe(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSubscribeUseCase_EmailServiceError(t *testing.T) {
	ts := setupSubscribeUseCaseTest()

	request := domain.SubscriptionRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupEmailErrorMocks(request)

	result, err := ts.useCase.Subscribe(context.Background(), request)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Token)
	assert.Contains(t, result.Message, "confirmation email failed")
}
