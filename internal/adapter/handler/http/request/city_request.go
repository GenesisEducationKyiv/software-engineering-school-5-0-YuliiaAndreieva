package request

import (
	"strings"
	"weather-api/internal/adapter/handler/http/errors"
)

type CityRequest struct {
	City string
}

func NewCityRequest(city string) *CityRequest {
	return &CityRequest{City: city}
}

func (r *CityRequest) Validate() error {
	if strings.TrimSpace(r.City) == "" {
		return errors.ErrCityRequired
	}

	return nil
}
