package domain

import "errors"

var (
	ErrCityNotFound           = errors.New("City not found")
	ErrInvalidInput           = errors.New("Invalid input")
	ErrEmailAlreadySubscribed = errors.New("Email already subscribed")
	ErrInvalidToken           = errors.New("Invalid token")
	ErrTokenNotFound          = errors.New("Token not found")
)

type Frequency string

const (
	FrequencyDaily  Frequency = "daily"
	FrequencyHourly Frequency = "hourly"
)

type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

type Subscription struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	City        string    `json:"city"`
	Frequency   Frequency `json:"frequency"`
	Token       string    `json:"token"`
	IsConfirmed bool      `json:"is_confirmed"`
}
