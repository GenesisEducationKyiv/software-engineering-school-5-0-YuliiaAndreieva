package unit

import (
	"context"
	"testing"
	"time"

	"weather-service/internal/core/domain"
	"weather-service/internal/core/usecase"
	"weather-service/internal/utils/logger"
	"weather-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usecaseTestSetup struct {
	useCase *usecase.GetWeatherUseCase
}

func setupUseCaseTest(t *testing.T, mockProvider *mocks.MockChainWeatherProvider) *usecaseTestSetup {
	logger := logger.NewLogrusLogger()
	useCase := usecase.NewGetWeatherUseCase(mockProvider, logger)

	return &usecaseTestSetup{
		useCase: useCase.(*usecase.GetWeatherUseCase),
	}
}

func (uts *usecaseTestSetup) makeRequest(t *testing.T, city string) (*domain.WeatherResponse, error) {
	request := domain.WeatherRequest{
		City: city,
	}

	return uts.useCase.GetWeather(context.Background(), request)
}

func TestGetWeatherUseCase_GetWeather_Success(t *testing.T) {
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

	ts := setupUseCaseTest(t, mockProvider)

	result, err := ts.makeRequest(t, "Kyiv")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, expectedWeather.City, result.Weather.City)
	assert.Equal(t, expectedWeather.Temperature, result.Weather.Temperature)
	assert.Equal(t, expectedWeather.Humidity, result.Weather.Humidity)
	assert.Equal(t, expectedWeather.Description, result.Weather.Description)
	assert.Equal(t, expectedWeather.WindSpeed, result.Weather.WindSpeed)
}

func TestGetWeatherUseCase_GetWeather_MultipleCities(t *testing.T) {
	cities := []string{"Kyiv", "Lviv", "Kharkiv"}

	for _, city := range cities {
		t.Run(city, func(t *testing.T) {
			expectedWeather := domain.Weather{
				City:        city,
				Temperature: 15.5,
				Humidity:    65,
				Description: "Partly cloudy",
				WindSpeed:   12.3,
				Timestamp:   time.Now(),
			}

			mockProvider := mocks.NewMockChainWeatherProvider().
				WithGetWeather(func(ctx context.Context, cityName string) (domain.Weather, error) {
					return expectedWeather, nil
				})

			ts := setupUseCaseTest(t, mockProvider)

			result, err := ts.makeRequest(t, city)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Success)
			assert.Equal(t, expectedWeather.City, result.Weather.City)
			assert.Equal(t, expectedWeather.Temperature, result.Weather.Temperature)
			assert.Equal(t, expectedWeather.Humidity, result.Weather.Humidity)
			assert.Equal(t, expectedWeather.Description, result.Weather.Description)
			assert.Equal(t, expectedWeather.WindSpeed, result.Weather.WindSpeed)
		})
	}
}

func TestGetWeatherUseCase_GetWeather_ProviderError(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			return domain.Weather{}, assert.AnError
		})

	ts := setupUseCaseTest(t, mockProvider)

	result, err := ts.makeRequest(t, "Kyiv")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to get weather data")
	assert.Contains(t, result.Error, "assert.AnError general error for testing")
}

func TestGetWeatherUseCase_GetWeather_ContextCancellation(t *testing.T) {
	mockProvider := mocks.NewMockChainWeatherProvider().
		WithGetWeather(func(ctx context.Context, city string) (domain.Weather, error) {
			select {
			case <-ctx.Done():
				return domain.Weather{}, ctx.Err()
			default:
				return domain.Weather{
					City:        city,
					Temperature: 15.5,
					Humidity:    65,
					Description: "Partly cloudy",
					WindSpeed:   12.3,
					Timestamp:   time.Now(),
				}, nil
			}
		})

	ts := setupUseCaseTest(t, mockProvider)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ts.useCase.GetWeather(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "Failed to get weather data")
	assert.Contains(t, result.Error, "context canceled")
}
