package in

import (
	"email/internal/adapter/dto"
	"email/internal/core/domain"
)

type EmailMapper interface {
	MapConfirmationEmailRequest(req dto.ConfirmationEmailRequest) domain.ConfirmationEmailRequest
	MapWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) domain.WeatherUpdateEmailRequest
}
