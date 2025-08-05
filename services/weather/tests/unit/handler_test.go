package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httphandler "weather/internal/adapter/http"
	"weather/internal/core/domain"
	"weather/internal/utils/logger"

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
		Weather: domain.Weather{},
	}, nil
}

type handlerTestSetup struct {
	handler *httphandler.WeatherHandler
	router  *gin.Engine
}

func setupHandlerTest(mockUseCase *MockGetWeatherUseCase) *handlerTestSetup {
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
				Weather: expectedWeather,
			}, nil
		},
	}

	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusOK, w.Code)
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
				Weather: expectedWeather,
			}, nil
		},
	}

	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "Lviv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedWeather.City, response.Weather.City)
	assert.Equal(t, expectedWeather.Temperature, response.Weather.Temperature)
	assert.Equal(t, expectedWeather.Humidity, response.Weather.Humidity)
	assert.Equal(t, expectedWeather.Description, response.Weather.Description)
	assert.Equal(t, expectedWeather.WindSpeed, response.Weather.WindSpeed)
}

func TestWeatherHandler_GetWeather_InvalidJSON(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(mockUseCase)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBufferString(`{"invalid": json`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response domain.WeatherResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request", response.Message)
}

func TestWeatherHandler_GetWeather_EmptyCity(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "City is required", response.Message)
}

func TestWeatherHandler_GetWeather_WhitespaceCity(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "   ",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "City is required", response.Message)
}

func TestWeatherHandler_GetWeather_UsecaseError(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return nil, assert.AnError
		},
	}

	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Failed to get weather data", response.Message)
}

func TestWeatherHandler_GetWeather_ProviderFailure(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{
		getWeatherFunc: func(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
			return &domain.WeatherResponse{
				Message: "Failed to get weather data",
				Error:   "provider error",
			}, nil
		},
	}

	ts := setupHandlerTest(mockUseCase)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Failed to get weather data", response.Message)
	assert.Equal(t, "provider error", response.Error)
}

func TestWeatherHandler_GetWeather_MissingContentType(t *testing.T) {
	mockUseCase := &MockGetWeatherUseCase{}
	ts := setupHandlerTest(mockUseCase)

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
