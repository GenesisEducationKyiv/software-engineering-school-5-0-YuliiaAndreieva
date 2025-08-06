package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httphandler "subscription/internal/adapter/http"
	"subscription/internal/config"
	"subscription/internal/core/domain"
	"subscription/internal/core/usecase"
	"subscription/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type subscriptionHandlerTestSetup struct {
	handler            *httphandler.SubscriptionHandler
	router             *gin.Engine
	mockRepo           *mocks.SubscriptionRepository
	mockTokenService   *mocks.TokenService
	mockEventPublisher *mocks.EventPublisher
	mockLogger         *mocks.Logger
}

func setupSubscriptionHandlerTest() *subscriptionHandlerTestSetup {
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

	subscribeUseCase := usecase.NewSubscribeUseCase(mockRepo, mockTokenService, mockEventPublisher, mockLogger, config)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(mockRepo, mockTokenService, mockLogger)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(mockRepo, mockTokenService, mockLogger)
	listByFrequencyUseCase := usecase.NewListByFrequencyUseCase(mockRepo, mockLogger)

	handler := httphandler.NewSubscriptionHandler(
		subscribeUseCase,
		confirmUseCase,
		unsubscribeUseCase,
		listByFrequencyUseCase,
		mockLogger,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subscribe", handler.Subscribe)
	router.GET("/confirm/:token", handler.Confirm)
	router.DELETE("/unsubscribe/:token", handler.Unsubscribe)
	router.GET("/list", handler.ListByFrequency)

	return &subscriptionHandlerTestSetup{
		handler:            handler,
		router:             router,
		mockRepo:           mockRepo,
		mockTokenService:   mockTokenService,
		mockEventPublisher: mockEventPublisher,
		mockLogger:         mockLogger,
	}
}

func (ts *subscriptionHandlerTestSetup) setupSuccessMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockEventPublisher.On("PublishSubscriptionCreated", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *subscriptionHandlerTestSetup) setupValidationMocks() {
	ts.mockTokenService.On("GenerateToken", mock.Anything, mock.Anything, mock.Anything).Return("test-token", nil)
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *subscriptionHandlerTestSetup) setupRepositoryErrorMocks(request domain.SubscriptionRequest) {
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(assert.AnError)
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *subscriptionHandlerTestSetup) setupJSONErrorMocks() {
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
}

func (ts *subscriptionHandlerTestSetup) makeSubscribeRequest(t *testing.T, request domain.SubscriptionRequest) (*httptest.ResponseRecorder, *domain.SubscriptionResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/subscribe", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	var response domain.SubscriptionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestSubscriptionHandler_Subscribe_Success(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Valid subscription request", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupSuccessMocks(request)

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "Subscription successful")

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockEventPublisher.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_Subscribe_InvalidEmail(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Invalid email format", func(t *testing.T) {
		ts.setupValidationMocks()
		ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()

		request := domain.SubscriptionRequest{
			Email:     "invalid-email",
			City:      "Kyiv",
			Frequency: "daily",
		}

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})
}

func TestSubscriptionHandler_Subscribe_EmptyFields(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Empty required fields", func(t *testing.T) {
		ts.setupValidationMocks()
		ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()

		request := domain.SubscriptionRequest{
			Email:     "",
			City:      "",
			Frequency: "",
		}

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)
	})
}

func TestSubscriptionHandler_Subscribe_UsecaseError(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Usecase error", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.setupRepositoryErrorMocks(request)

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Message)

		ts.mockRepo.AssertExpectations(t)
		ts.mockTokenService.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_Subscribe_DuplicateSubscription(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Duplicate subscription", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
		ts.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(domain.ErrDuplicateSubscription)
		ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
		ts.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
		ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
		ts.mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return()

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "already exists")

		ts.mockRepo.AssertExpectations(t)
		ts.mockLogger.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_InvalidJSON(t *testing.T) {
	ts := setupSubscriptionHandlerTest()

	t.Run("Invalid JSON", func(t *testing.T) {
		ts.setupJSONErrorMocks()

		req := httptest.NewRequest("POST", "/subscribe", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		ts.mockLogger.AssertExpectations(t)
	})
}
