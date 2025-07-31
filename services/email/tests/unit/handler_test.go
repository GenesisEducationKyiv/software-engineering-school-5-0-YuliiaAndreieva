package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"email/internal/adapter/dto"
	httphandler "email/internal/adapter/http"
	"email/internal/adapter/logger"
	"email/internal/core/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSendEmailUseCase struct {
	sendConfirmationEmailFunc  func(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error)
	sendWeatherUpdateEmailFunc func(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error)
}

func (m *MockSendEmailUseCase) SendConfirmationEmail(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
	if m.sendConfirmationEmailFunc != nil {
		return m.sendConfirmationEmailFunc(ctx, req)
	}
	return &domain.EmailDeliveryResult{
		To:     req.To,
		Status: domain.StatusDelivered,
	}, nil
}

func (m *MockSendEmailUseCase) SendWeatherUpdateEmail(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
	if m.sendWeatherUpdateEmailFunc != nil {
		return m.sendWeatherUpdateEmailFunc(ctx, req)
	}
	return &domain.EmailDeliveryResult{
		To:     req.To,
		Status: domain.StatusDelivered,
	}, nil
}

type emailHandlerTestSetup struct {
	handler *httphandler.EmailHandler
	router  *gin.Engine
}

func setupEmailHandlerTest(mockUseCase *MockSendEmailUseCase) *emailHandlerTestSetup {
	logger := logger.NewLogrusLogger()
	handler := httphandler.NewEmailHandler(mockUseCase, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/send/confirmation", handler.SendConfirmationEmail)
	router.POST("/send/weather-update", handler.SendWeatherUpdateEmail)

	return &emailHandlerTestSetup{
		handler: handler,
		router:  router,
	}
}

func (ehts *emailHandlerTestSetup) makeConfirmationRequest(t *testing.T, request dto.ConfirmationEmailRequest) (*httptest.ResponseRecorder, *domain.EmailResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ehts.router.ServeHTTP(w, req)

	var response domain.EmailResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func (ehts *emailHandlerTestSetup) makeWeatherUpdateRequest(t *testing.T, request dto.WeatherUpdateEmailRequest) (*httptest.ResponseRecorder, *domain.EmailResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/send/weather-update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ehts.router.ServeHTTP(w, req)

	var response domain.EmailResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestEmailHandler_SendConfirmationEmail_Success(t *testing.T) {
	tests := []struct {
		name            string
		request         dto.ConfirmationEmailRequest
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "Valid confirmation email request",
			request: dto.ConfirmationEmailRequest{
				To:               "test@example.com",
				Subject:          "Confirm Subscription",
				City:             "Kyiv",
				ConfirmationLink: "http://localhost/confirm/token123",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "Another valid confirmation email request",
			request: dto.ConfirmationEmailRequest{
				To:               "user@test.com",
				Subject:          "Confirm Weather Subscription",
				City:             "Lviv",
				ConfirmationLink: "http://localhost/confirm/token456",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &MockSendEmailUseCase{
				sendConfirmationEmailFunc: func(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
					return &domain.EmailDeliveryResult{
						To:     req.To,
						Status: domain.StatusDelivered,
					}, nil
				},
			}

			ts := setupEmailHandlerTest(mockUseCase)

			w, response := ts.makeConfirmationRequest(t, tt.request)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedSuccess, response.Success)
		})
	}
}

func TestEmailHandler_SendConfirmationEmail_InvalidEmail(t *testing.T) {
	request := dto.ConfirmationEmailRequest{
		To:               "invalid-email",
		Subject:          "Confirm Subscription",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token123",
	}

	mockUseCase := &MockSendEmailUseCase{
		sendConfirmationEmailFunc: func(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeConfirmationRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_SendConfirmationEmail_EmptyFields(t *testing.T) {
	request := dto.ConfirmationEmailRequest{
		To:               "",
		Subject:          "",
		City:             "",
		ConfirmationLink: "",
	}

	mockUseCase := &MockSendEmailUseCase{
		sendConfirmationEmailFunc: func(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeConfirmationRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_SendConfirmationEmail_UsecaseError(t *testing.T) {
	request := dto.ConfirmationEmailRequest{
		To:               "test@example.com",
		Subject:          "Confirm Subscription",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token123",
	}

	mockUseCase := &MockSendEmailUseCase{
		sendConfirmationEmailFunc: func(ctx context.Context, req domain.ConfirmationEmailRequest) (*domain.EmailDeliveryResult, error) {
			return nil, assert.AnError
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeConfirmationRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_SendWeatherUpdateEmail_Success(t *testing.T) {
	tests := []struct {
		name            string
		request         dto.WeatherUpdateEmailRequest
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "Valid weather update email request",
			request: dto.WeatherUpdateEmailRequest{
				To:          "test@example.com",
				Subject:     "Weather Update",
				Name:        "User",
				City:        "Kyiv",
				Temperature: 15,
				Description: "Partly cloudy",
				Humidity:    65,
				WindSpeed:   12,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "Another valid weather update email request",
			request: dto.WeatherUpdateEmailRequest{
				To:          "user@test.com",
				Subject:     "Daily Weather Report",
				Name:        "John",
				City:        "Lviv",
				Temperature: 22,
				Description: "Sunny",
				Humidity:    45,
				WindSpeed:   8,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &MockSendEmailUseCase{
				sendWeatherUpdateEmailFunc: func(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
					return &domain.EmailDeliveryResult{
						To:     req.To,
						Status: domain.StatusDelivered,
					}, nil
				},
			}

			ts := setupEmailHandlerTest(mockUseCase)

			w, response := ts.makeWeatherUpdateRequest(t, tt.request)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedSuccess, response.Success)
		})
	}
}

func TestEmailHandler_SendWeatherUpdateEmail_InvalidEmail(t *testing.T) {
	request := dto.WeatherUpdateEmailRequest{
		To:          "invalid-email",
		Subject:     "Weather Update",
		Name:        "User",
		City:        "Kyiv",
		Temperature: 15,
		Description: "Partly cloudy",
		Humidity:    65,
		WindSpeed:   12,
	}

	mockUseCase := &MockSendEmailUseCase{
		sendWeatherUpdateEmailFunc: func(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeWeatherUpdateRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_SendWeatherUpdateEmail_MissingFields(t *testing.T) {
	request := dto.WeatherUpdateEmailRequest{
		To:          "test@example.com",
		Subject:     "",
		Name:        "",
		City:        "",
		Temperature: 15,
		Description: "",
		Humidity:    65,
		WindSpeed:   12,
	}

	mockUseCase := &MockSendEmailUseCase{
		sendWeatherUpdateEmailFunc: func(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
			return &domain.EmailDeliveryResult{
				To:     req.To,
				Status: domain.StatusDelivered,
			}, nil
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeWeatherUpdateRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_SendWeatherUpdateEmail_UsecaseError(t *testing.T) {
	request := dto.WeatherUpdateEmailRequest{
		To:          "test@example.com",
		Subject:     "Weather Update",
		Name:        "User",
		City:        "Kyiv",
		Temperature: 15,
		Description: "Partly cloudy",
		Humidity:    65,
		WindSpeed:   12,
	}

	mockUseCase := &MockSendEmailUseCase{
		sendWeatherUpdateEmailFunc: func(ctx context.Context, req domain.WeatherUpdateEmailRequest) (*domain.EmailDeliveryResult, error) {
			return nil, assert.AnError
		},
	}

	ts := setupEmailHandlerTest(mockUseCase)

	w, response := ts.makeWeatherUpdateRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
}

func TestEmailHandler_InvalidJSON(t *testing.T) {
	mockUseCase := &MockSendEmailUseCase{}
	ts := setupEmailHandlerTest(mockUseCase)

	t.Run("Invalid JSON for confirmation email", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON for weather update email", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/send/weather-update", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEmailHandler_ContentTypeValidation(t *testing.T) {
	mockUseCase := &MockSendEmailUseCase{}
	ts := setupEmailHandlerTest(mockUseCase)

	t.Run("Missing Content-Type header", func(t *testing.T) {
		request := dto.ConfirmationEmailRequest{
			To:               "test@example.com",
			Subject:          "Test",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost/confirm/token",
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBuffer(jsonData))

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
