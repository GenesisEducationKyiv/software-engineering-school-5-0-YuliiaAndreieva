package usecase

import (
	"context"
	"weather-service/internal/core/domain"
	"weather-service/internal/core/ports/in"
	"weather-service/internal/core/ports/out"
)

type GetWeatherUseCase struct {
	weatherProvider out.WeatherProvider
}

func NewGetWeatherUseCase(weatherProvider out.WeatherProvider) in.GetWeatherUseCase {
	return &GetWeatherUseCase{
		weatherProvider: weatherProvider,
	}
}

func (uc *GetWeatherUseCase) GetWeather(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
	weather, err := uc.weatherProvider.GetWeather(ctx, req.City)
	if err != nil {
		return &domain.WeatherResponse{
			Success: false,
			Message: "Failed to get weather data",
			Error:   err.Error(),
		}, nil
	}

	return &domain.WeatherResponse{
		Success: true,
		Weather: weather,
		Message: "Weather data retrieved successfully",
	}, nil
}
