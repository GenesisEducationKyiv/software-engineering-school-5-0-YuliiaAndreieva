package http

import (
	"email-service/internal/adapter/dto"
	"strings"
)

type EmailValidator struct{}

func NewEmailValidator() *EmailValidator {
	return &EmailValidator{}
}

func (v *EmailValidator) ValidateEmailFormat(email string) bool {
	if email == "" {
		return false
	}
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (v *EmailValidator) ValidateRequiredFields(fields map[string]string) []string {
	var errors []string
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			errors = append(errors, field+" is required")
		}
	}
	return errors
}

func (v *EmailValidator) ValidateConfirmationEmailRequest(req dto.ConfirmationEmailRequest) []string {
	var errors []string

	if !v.ValidateEmailFormat(req.To) {
		errors = append(errors, "Invalid email format")
	}

	requiredFields := map[string]string{
		"To":               req.To,
		"Subject":          req.Subject,
		"City":             req.City,
		"ConfirmationLink": req.ConfirmationLink,
	}

	fieldErrors := v.ValidateRequiredFields(requiredFields)
	errors = append(errors, fieldErrors...)

	return errors
}

func (v *EmailValidator) ValidateWeatherUpdateEmailRequest(req dto.WeatherUpdateEmailRequest) []string {
	var errors []string

	if !v.ValidateEmailFormat(req.To) {
		errors = append(errors, "Invalid email format")
	}

	requiredFields := map[string]string{
		"To":          req.To,
		"Subject":     req.Subject,
		"Name":        req.Name,
		"City":        req.City,
		"Description": req.Description,
	}

	fieldErrors := v.ValidateRequiredFields(requiredFields)
	errors = append(errors, fieldErrors...)

	return errors
}
