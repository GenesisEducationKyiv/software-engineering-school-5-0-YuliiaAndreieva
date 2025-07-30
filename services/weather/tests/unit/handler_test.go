package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httphandler "weather-service/internal/adapter/http"
	"weather-service/internal/core/domain"
	"weather-service/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockGetWeatherUseCase struct {
	getWeatherFunc func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error)
}

func (m *MockGetWeatherUseCase) GetWeather(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
	if m.getWeatherFunc != nil {
		return m.getWeatherFunc(ctx, req)
	}
	return &domain.WeatherResponse{
		Success: true,
		Weather: domain.Weather{},
		Message: "Success",
	}, nil
}

type handlerTestSetup struct {
	handler *httphandler.WeatherHandler
	router  *gin.Engine
}

func setupHandlerTest(t *testing.T, mockUseCase *MockGetWeatherUseCase) *handlerTestSetup {
	logger := logger.NewLogrusLogger()
	handler := httphandler.NewWeatherHandler(mockUseCase, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/weather", handler.GetWeather)

	return &handlerTestSetup{
		handler: handler,
		router:  router,
	}
}

func (hts *handlerTestSetup) makeRequest(t *testing.T, request domain.WeatherRequest) (*httptest.ResponseRecorder, *domain.WeatherResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	hts.router.ServeHTTP(w, req)

	var response domain.WeatherResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestWeatherHandler_GetWeather_Success(t *testing.T) {
	expectedWeather := domain.Weather{
		City:        "Kyiv",
		Temperature: 15.5,
		Humidity:    65,
		Description: "Partly cloudy",
		WindSpeed:   12.3,
		Timestamp:   time.Now(),
	}

	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return &domain.WeatherResponse{
				Success: true,
				Weather: expectedWeather,
				Message: "Weather data retrieved successfully",
			}, nil
		},
	}

	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, response.Success)
	assert.Equal(t, expectedWeather.City, response.Weather.City)
	assert.Equal(t, expectedWeather.Temperature, response.Weather.Temperature)
	assert.Equal(t, expectedWeather.Humidity, response.Weather.Humidity)
	assert.Equal(t, expectedWeather.Description, response.Weather.Description)
	assert.Equal(t, expectedWeather.WindSpeed, response.Weather.WindSpeed)
}

func TestWeatherHandler_GetWeather_AnotherValidRequest(t *testing.T) {
	expectedWeather := domain.Weather{
		City:        "Lviv",
		Temperature: 18.0,
		Humidity:    70,
		Description: "Sunny",
		WindSpeed:   8.5,
		Timestamp:   time.Now(),
	}

	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return &domain.WeatherResponse{
				Success: true,
				Weather: expectedWeather,
				Message: "Weather data retrieved successfully",
			}, nil
		},
	}

	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "Lviv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, response.Success)
	assert.Equal(t, expectedWeather.City, response.Weather.City)
	assert.Equal(t, expectedWeather.Temperature, response.Weather.Temperature)
	assert.Equal(t, expectedWeather.Humidity, response.Weather.Humidity)
	assert.Equal(t, expectedWeather.Description, response.Weather.Description)
	assert.Equal(t, expectedWeather.WindSpeed, response.Weather.WindSpeed)
}

func TestWeatherHandler_GetWeather_InvalidJSON(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(t, mockUseCase)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBufferString(`{"invalid": json`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response domain.WeatherResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid request")
}

func TestWeatherHandler_GetWeather_EmptyCity(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "City is required")
}

func TestWeatherHandler_GetWeather_WhitespaceCity(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "   ",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "City is required")
}

func TestWeatherHandler_GetWeather_UsecaseError(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return nil, assert.AnError
		},
	}

	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to get weather data")
}

func TestWeatherHandler_GetWeather_ProviderFailure(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return &domain.WeatherResponse{
				Success: false,
				Message: "Failed to get weather data",
				Error:   "provider error",
			}, nil
		},
	}

	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to get weather data")
	assert.Contains(t, response.Error, "provider error")
}

func TestWeatherHandler_GetWeather_MissingContentType(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(t, mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonData))

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
