package request

import "weather-api/internal/core/domain"

type SubscribeRequest struct {
	Email     string           `json:"email"`
	City      string           `json:"city"`
	Frequency domain.Frequency `json:"frequency"`
}
