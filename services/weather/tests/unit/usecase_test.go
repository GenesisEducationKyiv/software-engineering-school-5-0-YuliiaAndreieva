package unit

import (
	"context"
	"testing"
	"time"

	"weather/internal/core/domain"
	"weather/internal/core/usecase"
	"weather/internal/utils/logger"
	"weather/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usecaseTestSetup struct {
	useCase *usecase.GetWeatherUseCase
}

func setupUseCaseTest(mockProvider *mocks.MockChainWeatherProvider) *usecaseTestSetup {
	logger := logger.NewLogrusLogger()
	useCase := usecase.NewGetWeatherUseCase(mockProvider, logger)

	return &usecaseTestSetup{
		useCase: useCase.(*usecase.GetWeatherUseCase),
	}
}

func (uts *usecaseTestSetup) makeRequest(city string) (*domain.WeatherResponse, error) {
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

	ts := setupUseCaseTest(mockProvider)

	result, err := ts.makeRequest("Kyiv")

	require.NoError(t, err)
	assert.NotNil(t, result)
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

			ts := setupUseCaseTest(mockProvider)

			result, err := ts.makeRequest(city)

			require.NoError(t, err)
			assert.NotNil(t, result)
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

	ts := setupUseCaseTest(mockProvider)

	result, err := ts.makeRequest("Kyiv")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError general error for testing")
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

	ts := setupUseCaseTest(mockProvider)

	request := domain.WeatherRequest{
		City: "Kyiv",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ts.useCase.GetWeather(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context canceled")
}
