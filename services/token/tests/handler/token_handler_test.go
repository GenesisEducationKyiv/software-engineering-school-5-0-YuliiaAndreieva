package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httphandler "token/internal/adapter/http"
	"token/internal/core/domain"
	"token/internal/core/usecase"
	"token/tests"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenHandlerTestSetup struct {
	handler    *httphandler.TokenHandler
	router     *gin.Engine
	mockLogger *tests.MockLogger
	validToken string
}

func setupTokenHandlerTest() *tokenHandlerTestSetup {
	mockLogger := &tests.MockLogger{}
	secret := "test-secret-key-for-jwt-signing"

	generateUseCase := usecase.NewGenerateTokenUseCase(mockLogger, secret)
	validateUseCase := usecase.NewValidateTokenUseCase(mockLogger, secret)

	handler := httphandler.NewTokenHandler(generateUseCase, validateUseCase, mockLogger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/generate", handler.GenerateToken)
	router.POST("/validate", handler.ValidateToken)

	tests.SetupCommonLoggerMocks(mockLogger)
	validToken, err := generateUseCase.GenerateToken(context.Background(), domain.GenerateTokenRequest{
		Subject:   "test@example.com",
		ExpiresIn: "24h",
	})
	if err != nil {
		panic("Failed to generate test token")
	}

	return &tokenHandlerTestSetup{
		handler:    handler,
		router:     router,
		mockLogger: mockLogger,
		validToken: validToken.Token,
	}
}

func (ts *tokenHandlerTestSetup) setupSuccessMocks() {
	tests.SetupSuccessLoggerMocks(ts.mockLogger)
}

func (ts *tokenHandlerTestSetup) setupErrorMocks() {
	tests.SetupErrorLoggerMocks(ts.mockLogger)
}

func (ts *tokenHandlerTestSetup) makeGenerateTokenRequest(t *testing.T, request domain.GenerateTokenRequest) (*httptest.ResponseRecorder, *domain.GenerateTokenResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	var response domain.GenerateTokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func (ts *tokenHandlerTestSetup) makeValidateTokenRequest(t *testing.T, request domain.ValidateTokenRequest) (*httptest.ResponseRecorder, *domain.ValidateTokenResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	var response domain.ValidateTokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestTokenHandler_GenerateToken_Success(t *testing.T) {
	ts := setupTokenHandlerTest()

	t.Run("Valid token generation request", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "test@example.com",
			ExpiresIn: "24h",
		}

		ts.setupSuccessMocks()

		w, response := ts.makeGenerateTokenRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "Token generated successfully")
	})

	t.Run("Another valid token generation request", func(t *testing.T) {
		request := domain.GenerateTokenRequest{
			Subject:   "user@test.com",
			ExpiresIn: "1h",
		}

		ts.setupSuccessMocks()

		w, response := ts.makeGenerateTokenRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "Token generated successfully")
	})
}

func TestTokenHandler_GenerateToken_InvalidJSON(t *testing.T) {
	ts := setupTokenHandlerTest()

	t.Run("Invalid JSON for token generation", func(t *testing.T) {
		ts.setupErrorMocks()

		req := httptest.NewRequest("POST", "/generate", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.GenerateTokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid request")
	})
}

func TestTokenHandler_ValidateToken_Success(t *testing.T) {
	ts := setupTokenHandlerTest()

	t.Run("Valid token validation", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: ts.validToken,
		}

		ts.setupSuccessMocks()

		w, response := ts.makeValidateTokenRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.True(t, response.Valid)
		assert.Contains(t, response.Message, "Token is valid")
	})
}

func TestTokenHandler_ValidateToken_InvalidToken(t *testing.T) {
	ts := setupTokenHandlerTest()

	t.Run("Invalid token", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: "invalid.token.here",
		}

		ts.setupSuccessMocks()

		w, response := ts.makeValidateTokenRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.False(t, response.Valid)
		assert.Contains(t, response.Message, "Token is invalid")
	})

	t.Run("Empty token", func(t *testing.T) {
		request := domain.ValidateTokenRequest{
			Token: "",
		}

		ts.setupSuccessMocks()

		w, response := ts.makeValidateTokenRequest(t, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, response.Success)
		assert.False(t, response.Valid)
		assert.Contains(t, response.Message, "Token is invalid")
	})
}

func TestTokenHandler_ValidateToken_InvalidJSON(t *testing.T) {
	ts := setupTokenHandlerTest()

	t.Run("Invalid JSON for token validation", func(t *testing.T) {
		ts.setupErrorMocks()

		req := httptest.NewRequest("POST", "/validate", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.ValidateTokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "Invalid request")
	})
}
