package in

import (
	"email-service/internal/adapter/dto"
	"email-service/internal/core/domain"
)

type EmailMapper interface {
	MapConfirmationEmailRequest(req dto.ConfirmationEmailRequest) domain.ConfirmationEmailRequest
	MapWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) domain.WeatherUpdateEmailRequest
}
