package integration

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
	"weather/internal/core/usecase"
	"weather/internal/utils/logger"
	"weather/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSetup struct {
	handler *httphandler.WeatherHandler
	router  *gin.Engine
}

func setupTestHandler(mockProvider *mocks.MockChainWeatherProvider) *testSetup {
	logrusLogger := logger.NewLogrusLogger()
	useCase := usecase.NewGetWeatherUseCase(mockProvider, logrusLogger)
	handler := httphandler.NewWeatherHandler(useCase, logrusLogger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/weather", handler.GetWeather)

	return &testSetup{
		handler: handler,
		router:  router,
	}
}

func (ts *testSetup) makeRequest(t *testing.T, request domain.WeatherRequest) (*httptest.ResponseRecorder, *domain.WeatherResponse) {
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	var response domain.WeatherResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return w, &response
}

func TestWeatherServiceIntegration_GetWeather_Success(t *testing.T) {
	expectedWeather := domain.Weather{
		City:        "Kyiv",
		Temperature: 15.5,
		Humidity:    65,
		Description: "Partly cloudy",
		WindSpeed:   12.3,
		Timestamp:   time.Now(),
	}

	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return expectedWeather, nil
		})

	ts := setupTestHandler(mockProvider)

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

func TestWeatherServiceIntegration_GetWeather_AnotherValidRequest(t *testing.T) {
	expectedWeather := domain.Weather{
		City:        "Lviv",
		Temperature: 18.0,
		Humidity:    70,
		Description: "Sunny",
		WindSpeed:   8.5,
		Timestamp:   time.Now(),
	}

	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return expectedWeather, nil
		})

	ts := setupTestHandler(mockProvider)

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

func TestWeatherServiceIntegration_GetWeather_EmptyCity(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider()
	ts := setupTestHandler(mockProvider)

	request := domain.WeatherRequest{
		City: "",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "City is required")
}

func TestWeatherServiceIntegration_GetWeather_WhitespaceCity(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider()
	ts := setupTestHandler(mockProvider)

	request := domain.WeatherRequest{
		City: "   ",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "City is required")
}

func TestWeatherServiceIntegration_GetWeather_ProviderFailure(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return domain.Weather{}, assert.AnError
		})

	ts := setupTestHandler(mockProvider)

	request := domain.WeatherRequest{
		City: "NonExistentCity123",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to get weather data")
}

func TestWeatherServiceIntegration_InvalidJSON(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider()
	ts := setupTestHandler(mockProvider)

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

func TestWeatherServiceIntegration_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "weather"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "weather", response["service"])
}

func TestWeatherServiceIntegration_WeatherProvider(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider().
		WithName(func() string {
			return "MockChainWeatherProvider"
		})

	t.Run("Get weather for valid city", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		expectedWeather := domain.Weather{
			City:        "Kyiv",
			Temperature: 15.5,
			Humidity:    65,
			Description: "Partly cloudy",
			WindSpeed:   12.3,
			Timestamp:   time.Now(),
		}

		mockProvider.WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return expectedWeather, nil
		})

		weather, err := mockProvider.GetWeather(ctx, "Kyiv")

		require.NoError(t, err)
		assert.NotNil(t, weather)
		assert.Equal(t, expectedWeather.City, weather.City)
		assert.Equal(t, expectedWeather.Temperature, weather.Temperature)
		assert.Equal(t, expectedWeather.Humidity, weather.Humidity)
		assert.Equal(t, expectedWeather.Description, weather.Description)
		assert.Equal(t, expectedWeather.WindSpeed, weather.WindSpeed)
	})

	t.Run("Check city exists", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mockProvider.WithCheckCityExists(func(ctx context.Context, city string) error {
			return nil
		})

		err := mockProvider.CheckCityExists(ctx, "Kyiv")

		assert.NoError(t, err)
	})

	t.Run("Provider name", func(t *testing.T) {
		name := mockProvider.Name()
		assert.Equal(t, "MockChainWeatherProvider", name)
	})
}

func TestWeatherServiceIntegration_ErrorHandling_ProviderFailure(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return domain.Weather{}, assert.AnError
		})

	ts := setupTestHandler(mockProvider)

	request := domain.WeatherRequest{
		City: "NonExistentCity123",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to get weather data")
}

func TestWeatherServiceIntegration_ErrorHandling_EmptyCity(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider()
	ts := setupTestHandler(mockProvider)

	request := domain.WeatherRequest{
		City: "",
	}

	w, response := ts.makeRequest(t, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "City is required")
}
