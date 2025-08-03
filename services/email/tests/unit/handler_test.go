package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"email/internal/adapter/dto"
	httphandler "email/internal/adapter/http"
	"email/internal/adapter/logger"
	"email/internal/core/domain"
	"email/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type emailHandlerTestSetup struct {
	handler     *httphandler.EmailHandler
	router      *gin.Engine
	mockUseCase *mocks.SendEmailUseCase
}

func setupEmailHandlerTest() *emailHandlerTestSetup {
	mockUseCase := &mocks.SendEmailUseCase{}
	logger := logger.NewLogrusLogger()
	handler := httphandler.NewEmailHandler(mockUseCase, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/send/confirmation", handler.SendConfirmationEmail)
	router.POST("/send/weather-update", handler.SendWeatherUpdateEmail)

	return &emailHandlerTestSetup{
		handler:     handler,
		router:      router,
		mockUseCase: mockUseCase,
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
			ehts := setupEmailHandlerTest()
			ehts.mockUseCase.On("SendConfirmationEmail", mock.Anything, mock.Anything).Return(&domain.EmailDeliveryResult{
				To:     tt.request.To,
				Status: domain.StatusDelivered,
			}, nil)

			w, response := ehts.makeConfirmationRequest(t, tt.request)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedSuccess, response.Success)
			ehts.mockUseCase.AssertExpectations(t)
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

	ehts := setupEmailHandlerTest()
	w, response := ehts.makeConfirmationRequest(t, request)

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

	ehts := setupEmailHandlerTest()

	w, response := ehts.makeConfirmationRequest(t, request)

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

	ehts := setupEmailHandlerTest()
	ehts.mockUseCase.On("SendConfirmationEmail", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w, response := ehts.makeConfirmationRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
	ehts.mockUseCase.AssertExpectations(t)
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
			ehts := setupEmailHandlerTest()
			ehts.mockUseCase.On("SendWeatherUpdateEmail", mock.Anything, mock.Anything).Return(&domain.EmailDeliveryResult{
				To:     tt.request.To,
				Status: domain.StatusDelivered,
			}, nil)

			w, response := ehts.makeWeatherUpdateRequest(t, tt.request)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedSuccess, response.Success)
			ehts.mockUseCase.AssertExpectations(t)
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

	ehts := setupEmailHandlerTest()
	w, response := ehts.makeWeatherUpdateRequest(t, request)

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

	ehts := setupEmailHandlerTest()

	w, response := ehts.makeWeatherUpdateRequest(t, request)

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

	ehts := setupEmailHandlerTest()
	ehts.mockUseCase.On("SendWeatherUpdateEmail", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w, response := ehts.makeWeatherUpdateRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Message)
	ehts.mockUseCase.AssertExpectations(t)
}

func TestEmailHandler_InvalidJSON(t *testing.T) {
	ehts := setupEmailHandlerTest()

	t.Run("Invalid JSON for confirmation email", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ehts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON for weather update email", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/send/weather-update", bytes.NewBufferString(`{"invalid": json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ehts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEmailHandler_ContentTypeValidation(t *testing.T) {
	ehts := setupEmailHandlerTest()

	t.Run("Missing Content-Type header", func(t *testing.T) {
		request := dto.ConfirmationEmailRequest{
			To:               "test@example.com",
			Subject:          "Test",
			City:             "Kyiv",
			ConfirmationLink: "http://localhost/confirm/token",
		}

		ehts.mockUseCase.On("SendConfirmationEmail", mock.Anything, mock.Anything).Return(&domain.EmailDeliveryResult{
			To:     request.To,
			Status: domain.StatusDelivered,
		}, nil)

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/send/confirmation", bytes.NewBuffer(jsonData))

		w := httptest.NewRecorder()
		ehts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		ehts.mockUseCase.AssertExpectations(t)
	})
}
