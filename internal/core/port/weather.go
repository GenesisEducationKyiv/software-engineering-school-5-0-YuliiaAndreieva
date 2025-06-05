package port

import "weather-api/internal/core/domain"

type WeatherService interface {
	GetWeather(city string) (domain.Weather, error)
}
