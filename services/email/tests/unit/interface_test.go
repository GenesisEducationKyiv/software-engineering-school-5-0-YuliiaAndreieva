package unit

import (
	"testing"

	"email/internal/adapter/dto"
	"email/internal/adapter/http"
	"email/internal/adapter/mappings"
	"email/internal/core/ports/in"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidator_ImplementsInterface(t *testing.T) {
	var _ in.EmailValidator = (*http.EmailValidator)(nil)

	validator := http.NewEmailValidator()
	assert.Implements(t, (*in.EmailValidator)(nil), validator)
}

func TestEmailMapper_ImplementsInterface(t *testing.T) {
	var _ in.EmailMapper = (*mappings.EmailMapper)(nil)

	mapper := mappings.NewEmailMapper()
	assert.Implements(t, (*in.EmailMapper)(nil), mapper)
}

func TestEmailValidator_InterfaceMethods(t *testing.T) {
	validator := http.NewEmailValidator()

	assert.IsType(t, true, validator.ValidateEmailFormat("test@example.com"))
	assert.IsType(t, []string{}, validator.ValidateRequiredFields(map[string]string{}))
	assert.IsType(t, []string{}, validator.ValidateConfirmationEmailRequest(dto.ConfirmationEmailRequest{}))
	assert.IsType(t, []string{}, validator.ValidateWeatherUpdateEmailRequest(dto.WeatherUpdateEmailRequest{}))
}

func TestEmailMapper_InterfaceMethods(t *testing.T) {
	mapper := mappings.NewEmailMapper()

	confirmationReq := dto.ConfirmationEmailRequest{
		To:               "test@example.com",
		Subject:          "Test",
		City:             "Kyiv",
		ConfirmationLink: "http://localhost/confirm/token",
	}

	weatherReq := dto.WeatherUpdateEmailRequest{
		To:          "test@example.com",
		Subject:     "Test",
		Name:        "User",
		City:        "Kyiv",
		Temperature: 15,
		Description: "Clear",
		Humidity:    50,
		WindSpeed:   10,
	}

	confirmationResult := mapper.MapConfirmationEmailRequest(confirmationReq)
	weatherResult := mapper.MapWeatherUpdateEmailRequest(weatherReq)

	assert.NotNil(t, confirmationResult)
	assert.NotNil(t, weatherResult)
}
