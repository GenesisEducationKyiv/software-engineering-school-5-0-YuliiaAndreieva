package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httphandler "subscription-service/internal/adapter/http"
	"subscription-service/internal/core/domain"
	"subscription-service/internal/core/ports/out"
	"subscription-service/internal/core/usecase"
	"subscription-service/tests"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type subscriptionHandlerTestSetup struct {
	handler          *httphandler.SubscriptionHandler
	router           *gin.Engine
	mockRepo         *tests.MockSubscriptionRepository
	mockTokenService *tests.MockTokenService
	mockEmailService *tests.MockEmailService
	mockLogger       *tests.MockLogger
}

func setupSubscriptionHandlerTest(t *testing.T) *subscriptionHandlerTestSetup {
	mockRepo := &tests.MockSubscriptionRepository{}
	mockTokenService := &tests.MockTokenService{}
	mockEmailService := &tests.MockEmailService{}
	mockLogger := &tests.MockLogger{}

	subscribeUseCase := usecase.NewSubscribeUseCase(mockRepo, mockTokenService, mockEmailService, mockLogger)
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
		handler:          handler,
		router:           router,
		mockRepo:         mockRepo,
		mockTokenService: mockTokenService,
		mockEmailService: mockEmailService,
		mockLogger:       mockLogger,
	}
}

func (ts *subscriptionHandlerTestSetup) setupSuccessMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
	ts.mockRepo.On("CreateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil)
	ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(out.EmailDeliveryResult{}, nil)
	tests.SetupSuccessLoggerMocks(ts.mockLogger)
}

func (ts *subscriptionHandlerTestSetup) setupValidationMocks() {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	ts.mockTokenService.On("GenerateToken", mock.Anything, mock.Anything, mock.Anything).Return("test-token", nil)
	tests.SetupCommonLoggerMocks(ts.mockLogger)
}

func (ts *subscriptionHandlerTestSetup) setupRepositoryErrorMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(false, assert.AnError)
	tests.SetupErrorLoggerMocks(ts.mockLogger)
}

func (ts *subscriptionHandlerTestSetup) setupDuplicateMocks(request domain.SubscriptionRequest) {
	ts.mockRepo.On("IsSubscriptionExists", mock.Anything, request.Email, request.City).Return(true, nil)
	tests.SetupWarningLoggerMocks(ts.mockLogger)
}

func (ts *subscriptionHandlerTestSetup) setupJSONErrorMocks() {
	tests.SetupJSONErrorLoggerMocks(ts.mockLogger)
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
	ts := setupSubscriptionHandlerTest(t)

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
		ts.mockEmailService.AssertExpectations(t)
	})

	t.Run("Another valid subscription request", func(t *testing.T) {
		request := domain.SubscriptionRequest{
			Email:     "user@test.com",
			City:      "Lviv",
			Frequency: "weekly",
		}

		ts.setupSuccessMocks(request)

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "Subscription successful")
	})
}

func TestSubscriptionHandler_Subscribe_InvalidEmail(t *testing.T) {
	ts := setupSubscriptionHandlerTest(t)

	request := domain.SubscriptionRequest{
		Email:     "invalid-email",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupValidationMocks()

	w, response := ts.makeSubscribeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestSubscriptionHandler_Subscribe_EmptyFields(t *testing.T) {
	ts := setupSubscriptionHandlerTest(t)

	request := domain.SubscriptionRequest{
		Email:     "",
		City:      "",
		Frequency: "",
	}

	ts.setupValidationMocks()

	w, response := ts.makeSubscribeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestSubscriptionHandler_Subscribe_UsecaseError(t *testing.T) {
	ts := setupSubscriptionHandlerTest(t)

	request := domain.SubscriptionRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupRepositoryErrorMocks(request)

	w, response := ts.makeSubscribeRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to process subscription")
}

func TestSubscriptionHandler_Subscribe_DuplicateSubscription(t *testing.T) {
	ts := setupSubscriptionHandlerTest(t)

	request := domain.SubscriptionRequest{
		Email:     "duplicate@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	ts.setupDuplicateMocks(request)

	w, response := ts.makeSubscribeRequest(t, request)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "already subscribed")
}

func TestSubscriptionHandler_InvalidJSON(t *testing.T) {
	ts := setupSubscriptionHandlerTest(t)

	t.Run("Invalid JSON for subscription", func(t *testing.T) {
		ts.setupJSONErrorMocks()

		req := httptest.NewRequest("POST", "/subscribe", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.SubscriptionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid request")
	})
}
