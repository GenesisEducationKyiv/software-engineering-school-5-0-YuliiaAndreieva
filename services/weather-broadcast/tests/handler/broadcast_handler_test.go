package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httphandler "weather-broadcast/internal/adapter/http"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBroadcastUseCase struct {
	mock.Mock
}

func (m *MockBroadcastUseCase) Broadcast(ctx context.Context, frequency domain.Frequency) error {
	args := m.Called(ctx, frequency)
	return args.Error(0)
}

type broadcastHandlerTestSetup struct {
	handler              *httphandler.BroadcastHandler
	router               *gin.Engine
	mockBroadcastUseCase *MockBroadcastUseCase
	mockLogger           *mocks.Logger
}

func setupBroadcastHandlerTest() *broadcastHandlerTestSetup {
	mockBroadcastUseCase := &MockBroadcastUseCase{}
	mockLogger := &mocks.Logger{}

	handler := httphandler.NewBroadcastHandler(mockBroadcastUseCase, mockLogger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/broadcast", handler.Broadcast)

	return &broadcastHandlerTestSetup{
		handler:              handler,
		router:               router,
		mockBroadcastUseCase: mockBroadcastUseCase,
		mockLogger:           mockLogger,
	}
}

func (ts *broadcastHandlerTestSetup) setupSuccessMocks() {
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()
}

func (ts *broadcastHandlerTestSetup) setupErrorMocks() {
	ts.mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return()
	ts.mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
}

func (ts *broadcastHandlerTestSetup) makeBroadcastRequest(t *testing.T, request domain.BroadcastRequest) (*httptest.ResponseRecorder, map[string]interface{}) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/broadcast", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return w, response
}

func TestBroadcastHandler_Success(t *testing.T) {
	ts := setupBroadcastHandlerTest()

	t.Run("Valid broadcast request", func(t *testing.T) {
		request := domain.BroadcastRequest{
			Frequency: domain.Daily,
		}

		ts.mockBroadcastUseCase.On("Broadcast", mock.Anything, domain.Daily).Return(nil)

		ts.setupSuccessMocks()

		w, response := ts.makeBroadcastRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Broadcast completed successfully")
	})

	t.Run("Another valid broadcast request", func(t *testing.T) {
		request := domain.BroadcastRequest{
			Frequency: domain.Weekly,
		}

		ts.mockBroadcastUseCase.On("Broadcast", mock.Anything, domain.Weekly).Return(nil)

		ts.setupSuccessMocks()

		w, response := ts.makeBroadcastRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Broadcast completed successfully")
	})
}

func TestBroadcastHandler_InvalidJSON(t *testing.T) {
	ts := setupBroadcastHandlerTest()

	t.Run("Invalid JSON for broadcast", func(t *testing.T) {
		ts.setupErrorMocks()

		req := httptest.NewRequest("POST", "/broadcast", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Invalid request")
	})
}

func TestBroadcastHandler_MissingFrequency(t *testing.T) {
	ts := setupBroadcastHandlerTest()

	t.Run("Missing frequency field", func(t *testing.T) {
		ts.setupErrorMocks()

		req := httptest.NewRequest("POST", "/broadcast", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Invalid request")
	})
}

func TestBroadcastHandler_BroadcastError(t *testing.T) {
	ts := setupBroadcastHandlerTest()

	t.Run("Broadcast usecase error", func(t *testing.T) {
		request := domain.BroadcastRequest{
			Frequency: domain.Daily,
		}

		ts.mockBroadcastUseCase.On("Broadcast", mock.Anything, domain.Daily).Return(assert.AnError)

		ts.setupErrorMocks()

		w, response := ts.makeBroadcastRequest(t, request)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Broadcast failed")
	})
}
