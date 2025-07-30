package in

import "email/internal/adapter/dto"

type EmailValidator interface {
	ValidateEmailFormat(email string) bool
	ValidateRequiredFields(fields map[string]string) []string
	ValidateConfirmationEmailRequest(req dto.ConfirmationEmailRequest) []string
	ValidateWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) []string
}
