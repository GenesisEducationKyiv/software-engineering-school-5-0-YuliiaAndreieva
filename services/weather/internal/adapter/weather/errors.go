package weather

import (
	"fmt"
)

type ProviderError struct {
	Provider string
	Code     int
	Message  string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s error (code: %d): %s", e.Provider, e.Code, e.Message)
}

func NewProviderError(provider string, code int, message string) error {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
	}
}
