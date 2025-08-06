package out

import (
	"context"
	"weather-api/internal/core/domain"
)

type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
	CheckCityExists(ctx context.Context, city string) error
	Name() string
}
