package request

import (
	"strings"
	"weather-api/internal/core/domain"
	"weather-api/internal/handler/http/errors"
)

type SubscribeRequest struct {
	Email     string           `json:"email"`
	City      string           `json:"city"`
	Frequency domain.Frequency `json:"frequency"`
}

func (r *SubscribeRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.ErrEmailRequired
	}

	if !isValidEmail(r.Email) {
		return errors.ErrInvalidEmail
	}

	if strings.TrimSpace(r.City) == "" {
		return errors.ErrCityRequired
	}

	if r.Frequency != domain.FrequencyDaily && r.Frequency != domain.FrequencyHourly {
		return errors.ErrInvalidFrequency
	}

	return nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
