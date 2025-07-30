package mappings

import (
	"email-service/internal/adapter/dto"
	"email-service/internal/core/domain"
)

type EmailMapper struct{}

func NewEmailMapper() *EmailMapper {
	return &EmailMapper{}
}

func (m *EmailMapper) MapConfirmationEmailRequest(req dto.ConfirmationEmailRequest) domain.ConfirmationEmailRequest {
	return domain.ConfirmationEmailRequest{
		To:               req.To,
		Subject:          req.Subject,
		City:             req.City,
		ConfirmationLink: req.ConfirmationLink,
	}
}

func (m *EmailMapper) MapWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) domain.WeatherUpdateEmailRequest {
	return domain.WeatherUpdateEmailRequest{
		To:          req.To,
		Subject:     req.Subject,
		Name:        req.Name,
		City:        req.City,
		Temperature: req.Temperature,
		Description: req.Description,
		Humidity:    req.Humidity,
		WindSpeed:   req.WindSpeed,
	}
}
