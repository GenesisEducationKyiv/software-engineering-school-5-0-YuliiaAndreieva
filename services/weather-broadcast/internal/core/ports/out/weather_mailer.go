package out

import (
	"context"
	"weather-broadcast/internal/core/domain"
)

//go:generate mockery --name WeatherMailer
type WeatherMailer interface {
	SendWeather(ctx context.Context, info *domain.WeatherMailSuccessInfo) error
	SendError(ctx context.Context, info *domain.WeatherMailErrorInfo) error
}
