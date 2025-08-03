package usecase

import (
	"context"
	"weather/internal/core/domain"
	"weather/internal/core/ports/in"
	"weather/internal/core/ports/out"
)

type GetWeatherUseCase struct {
	weatherProvider out.WeatherProvider
	logger          out.Logger
}

func NewGetWeatherUseCase(weatherProvider out.WeatherProvider, logger out.Logger) in.GetWeatherUseCase {
	return &GetWeatherUseCase{
		weatherProvider: weatherProvider,
		logger:          logger,
	}
}

func (uc *GetWeatherUseCase) GetWeather(ctx context.Context, req domain.WeatherRequest) (*domain.WeatherResponse, error) {
	uc.logger.Infof("Starting weather request for city: %s", req.City)

	uc.logger.Debugf("Fetching weather data from provider for city: %s", req.City)
	weather, err := uc.weatherProvider.GetWeather(ctx, req.City)
	if err != nil {
		uc.logger.Errorf("Failed to get weather data for city %s: %v", req.City, err)
		return &domain.WeatherResponse{
			Success: false,
			Message: "Failed to get weather data",
			Error:   err.Error(),
		}, nil
	}

	weather.City = req.City

	uc.logger.Infof("Successfully retrieved weather data for city: %s", req.City)
	uc.logger.Debugf("Weather data retrieved: temperature=%s, description=%s", weather.Temperature, weather.Description)

	return &domain.WeatherResponse{
		Success: true,
		Weather: weather,
		Message: "Weather data retrieved successfully",
	}, nil
}
