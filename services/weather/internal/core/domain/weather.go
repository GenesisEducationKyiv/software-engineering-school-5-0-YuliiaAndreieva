package domain

import "time"

type Weather struct {
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	Description string    `json:"description"`
	WindSpeed   float64   `json:"wind_speed"`
	Timestamp   time.Time `json:"timestamp"`
}

type WeatherRequest struct {
	City string `json:"city" validate:"required"`
}

type WeatherResponse struct {
	Success bool    `json:"success"`
	Weather Weather `json:"weather,omitempty"`
	Message string  `json:"message"`
	Error   string  `json:"error,omitempty"`
}
