package domain

import (
	"errors"
	"fmt"
)

var (
	ErrCityNotFound           = errors.New("city not found")
	ErrEmailAlreadySubscribed = errors.New("email already subscribed")
	ErrInvalidToken           = errors.New("invalid token")
	ErrTokenNotFound          = errors.New("token not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}
