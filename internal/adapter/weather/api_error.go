package weather

import (
	"fmt"
	"weather-api/internal/core/domain"
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("weatherapi: %s (code %d)", e.Message, e.Code)
}

func newAPIError(code int, msg string) error {
	return &APIError{Code: code, Message: msg}
}

func mapAPIError(e *APIError) error {
	if e == nil || e.Code == 0 {
		return nil
	}
	if e.Code == 1006 {
		return domain.ErrCityNotFound
	}
	return &APIError{Code: e.Code, Message: e.Message}
}
