package openweathermap

import (
	"log"
	"weather-api/internal/adapter/weather"
)

const (
	CityNotFound       = 404
	Unauthorized       = 401
	RateLimitExceeded  = 429
	InternalError      = 500
	BadGateway         = 502
	ServiceUnavailable = 503
	GatewayTimeout     = 504
)

type OpenWeatherErrorInfo struct {
	Message string
}

var openWeatherMapErrors = map[int]OpenWeatherErrorInfo{
	CityNotFound:       {Message: "City not found"},
	Unauthorized:       {Message: "Authentication failed"},
	RateLimitExceeded:  {Message: "Service temporarily unavailable - rate limit exceeded"},
	InternalError:      {Message: "Weather service temporarily unavailable"},
	BadGateway:         {Message: "Weather service temporarily unavailable"},
	ServiceUnavailable: {Message: "Weather service temporarily unavailable"},
	GatewayTimeout:     {Message: "Weather service temporarily unavailable"},
}

func (c *Client) mapError(code int, message string) error {
	if errorInfo, exists := openWeatherMapErrors[code]; exists {
		log.Printf("OpenWeatherMap error - Code: %d, Original: %s", code, message)
		return weather.NewProviderError(c.Name(), code, errorInfo.Message)
	}

	log.Printf("Unknown OpenWeatherMap error - Code: %d, Message: %s", code, message)
	return weather.NewProviderError(c.Name(), code, "Weather service temporarily unavailable")
}
