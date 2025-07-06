package domain

import (
	"time"
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

type City struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Subscription struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	CityID      int64     `json:"city_id"`
	City        *City     `json:"city,omitempty"`
	Frequency   Frequency `json:"frequency"`
	Token       string    `json:"token"`
	IsConfirmed bool      `json:"is_confirmed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WeatherUpdate struct {
	Subscription Subscription
	Weather      Weather
}
