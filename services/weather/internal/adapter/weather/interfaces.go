package weather

import (
	"context"
	"net/http"
	"weather/internal/core/domain"
)

type ProviderLogger interface {
	Log(providerName string, responseBody []byte)
}

type Provider interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
	CheckCityExists(ctx context.Context, city string) error
	Name() string
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
