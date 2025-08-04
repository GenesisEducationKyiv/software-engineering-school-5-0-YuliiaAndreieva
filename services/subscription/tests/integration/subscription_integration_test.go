package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"subscription/internal/core/usecase"
	"testing"

	"github.com/joho/godotenv"

	"subscription/internal/adapter/database"
	httphandler "subscription/internal/adapter/http"
	"subscription/internal/core/domain"
	"subscription/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"subscription/internal/config"
)

func init() {
	godotenv.Load("../../test.env")
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type subscriptionIntegrationTestSetup struct {
	handler            *httphandler.SubscriptionHandler
	router             *gin.Engine
	gormDB             *gorm.DB
	mockTokenService   *mocks.TokenService
	mockEmailService   *mocks.EmailService
	mockEventPublisher *mocks.EventPublisher
	mockLogger         *mocks.Logger
}

func setupSubscriptionIntegrationTest(t *testing.T) *subscriptionIntegrationTestSetup {
	dsn := "host=" + getEnvWithDefault("TEST_DB_HOST", "postgres-subscription") +
		" user=" + getEnvWithDefault("TEST_DB_USER", "test") +
		" password=" + getEnvWithDefault("TEST_DB_PASSWORD", "test") +
		" dbname=" + getEnvWithDefault("TEST_DB_NAME", "subscription_test") +
		" port=" + getEnvWithDefault("TEST_DB_PORT", "5432") +
		" sslmode=disable"

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = gormDB.AutoMigrate(&database.Subscription{})
	require.NoError(t, err)

	mockTokenService := &mocks.TokenService{}
	mockEmailService := &mocks.EmailService{}
	mockEventPublisher := &mocks.EventPublisher{}
	mockLogger := &mocks.Logger{}

	repository := database.NewSubscriptionRepo(gormDB, mockLogger)

	cfg := &config.Config{
		Token: config.TokenConfig{
			Expiration: "24h",
		},
	}

	subscribeUseCase := usecase.NewSubscribeUseCase(repository, mockTokenService, mockEmailService, mockEventPublisher, mockLogger, cfg)
	confirmUseCase := usecase.NewConfirmSubscriptionUseCase(repository, mockTokenService, mockLogger)
	unsubscribeUseCase := usecase.NewUnsubscribeUseCase(repository, mockTokenService, mockLogger)
	listByFrequencyUseCase := usecase.NewListByFrequencyUseCase(repository, mockLogger)

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

	return &subscriptionIntegrationTestSetup{
		handler:            handler,
		router:             router,
		gormDB:             gormDB,
		mockTokenService:   mockTokenService,
		mockEmailService:   mockEmailService,
		mockEventPublisher: mockEventPublisher,
		mockLogger:         mockLogger,
	}
}

func (sits *subscriptionIntegrationTestSetup) cleanup() {
	sits.gormDB.Exec("DELETE FROM subscriptions")
	sits.mockTokenService.ExpectedCalls = nil
	sits.mockEmailService.ExpectedCalls = nil
	sits.mockLogger.ExpectedCalls = nil
}

func (sits *subscriptionIntegrationTestSetup) setupLoggerMocks() {
	sits.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sits.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
	sits.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	sits.mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return()
	sits.mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
}

func (sits *subscriptionIntegrationTestSetup) setupEventPublisherMocks() {
	sits.mockEventPublisher.On("PublishSubscriptionCreated", mock.Anything, mock.Anything).Return(nil)
}

func (sits *subscriptionIntegrationTestSetup) makeSubscribeRequest(t *testing.T, request domain.SubscriptionRequest) (*httptest.ResponseRecorder, *domain.SubscriptionResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/subscribe", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	sits.router.ServeHTTP(w, req)

	var response domain.SubscriptionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func (sits *subscriptionIntegrationTestSetup) makeConfirmRequest(t *testing.T, token string) (*httptest.ResponseRecorder, *domain.ConfirmResponse) {
	req := httptest.NewRequest("GET", "/confirm/"+token, nil)

	w := httptest.NewRecorder()
	sits.router.ServeHTTP(w, req)

	var response domain.ConfirmResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func (sits *subscriptionIntegrationTestSetup) makeUnsubscribeRequest(t *testing.T, token string) (*httptest.ResponseRecorder, *domain.UnsubscribeResponse) {
	req := httptest.NewRequest("DELETE", "/unsubscribe/"+token, nil)

	w := httptest.NewRecorder()
	sits.router.ServeHTTP(w, req)

	var response domain.UnsubscribeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestSubscriptionIntegration_Subscribe(t *testing.T) {
	ts := setupSubscriptionIntegrationTest(t)
	defer ts.cleanup()

	t.Run("Valid subscription request", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("test-token", nil)
		ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(nil)

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "Subscription successful")
	})

	t.Run("Invalid email format", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "invalid-email",
			City:      "Kyiv",
			Frequency: "daily",
		}

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid email format")
	})

	t.Run("Empty required fields", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "",
			City:      "",
			Frequency: "",
		}

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "validation failed")
	})

	t.Run("Duplicate subscription", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "duplicate@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("duplicate-token", nil)
		ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(nil)

		w, response := ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)

		w, response = ts.makeSubscribeRequest(t, request)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "already exists")
	})
}

func TestSubscriptionIntegration_Confirm(t *testing.T) {
	ts := setupSubscriptionIntegrationTest(t)
	defer ts.cleanup()

	t.Run("Valid confirmation", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "confirm@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("confirm-token", nil)
		ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(nil)

		w, _ := ts.makeSubscribeRequest(t, request)
		assert.Equal(t, http.StatusOK, w.Code)

		ts.mockTokenService.On("ValidateToken", mock.Anything, "confirm-token").Return(true, nil)

		w, response := ts.makeConfirmRequest(t, "confirm-token")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Subscription confirmed")
	})

	t.Run("Invalid token", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		ts.mockTokenService.On("ValidateToken", mock.Anything, "invalid-token").Return(false, nil)

		w, response := ts.makeConfirmRequest(t, "invalid-token")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid token")
	})

	t.Run("Empty token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/confirm/", nil)
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSubscriptionIntegration_Unsubscribe(t *testing.T) {
	ts := setupSubscriptionIntegrationTest(t)
	defer ts.cleanup()

	t.Run("Valid unsubscribe", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		request := domain.SubscriptionRequest{
			Email:     "unsubscribe@example.com",
			City:      "Kyiv",
			Frequency: "daily",
		}

		ts.mockTokenService.On("GenerateToken", mock.Anything, request.Email, "24h").Return("unsubscribe-token", nil)
		ts.mockEmailService.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("domain.ConfirmationEmailRequest")).Return(nil)

		w, _ := ts.makeSubscribeRequest(t, request)
		assert.Equal(t, http.StatusOK, w.Code)

		ts.mockTokenService.On("ValidateToken", mock.Anything, "unsubscribe-token").Return(true, nil)

		w, _ = ts.makeConfirmRequest(t, "unsubscribe-token")
		assert.Equal(t, http.StatusOK, w.Code)

		w, response := ts.makeUnsubscribeRequest(t, "unsubscribe-token")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Successfully unsubscribed")
	})

	t.Run("Invalid token", func(t *testing.T) {
		ts.setupLoggerMocks()
		ts.setupEventPublisherMocks()

		ts.mockTokenService.On("ValidateToken", mock.Anything, "invalid-token").Return(false, nil)

		w, response := ts.makeUnsubscribeRequest(t, "invalid-token")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid token")
	})
}
