package in

import "email-service/internal/adapter/dto"

type EmailValidator interface {
	ValidateEmailFormat(email string) bool
	ValidateRequiredFields(fields map[string]string) []string
	ValidateConfirmationEmailRequest(req dto.ConfirmationEmailRequest) []string
	ValidateWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) []string
}
