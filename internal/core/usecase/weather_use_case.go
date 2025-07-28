package usecase

import (
	"context"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/ports/out"
)

type WeatherUseCase struct {
	weatherProvider out.WeatherProvider
}

func NewWeatherUseCase(
	weatherProvider out.WeatherProvider,
) *WeatherUseCase {
	return &WeatherUseCase{
		weatherProvider: weatherProvider,
	}
}

func (uc *WeatherUseCase) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	return uc.weatherProvider.GetWeather(ctx, city)
}
