package errors

import "errors"

var (
	ErrInvalidInput           = errors.New("invalid input")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidFrequency       = errors.New("invalid frequency")
	ErrCityRequired           = errors.New("city parameter is required")
	ErrEmailRequired          = errors.New("email is required")
	ErrTokenRequired          = errors.New("token is required")
	ErrCityNotFound           = errors.New("city not found")
	ErrEmailAlreadySubscribed = errors.New("email already subscribed")
	ErrTokenNotFound          = errors.New("token not found")
	ErrInvalidToken           = errors.New("invalid token")
)
