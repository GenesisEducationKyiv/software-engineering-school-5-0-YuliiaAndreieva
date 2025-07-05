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
	Temperature float64
	Humidity    int
	Description string
}

type City struct {
	ID   int64
	Name string
}

type Subscription struct {
	ID          int64
	Email       string
	CityID      int64
	City        *City
	Frequency   Frequency
	Token       string
	IsConfirmed bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WeatherUpdate struct {
	Subscription Subscription
	Weather      Weather
}
