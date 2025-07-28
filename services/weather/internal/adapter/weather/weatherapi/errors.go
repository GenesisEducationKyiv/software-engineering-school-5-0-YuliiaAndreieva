package weatherapi

import (
	"log"
	"weather-service/internal/adapter/weather"
)

const (
	KeyNotProvided      = 1002
	LocationNotProvided = 1003
	InvalidURL          = 1005
	LocationNotFound    = 1006
	InvalidKey          = 2006
	QuotaExceeded       = 2007
	KeyDisabled         = 2008
	NoAccess            = 2009
	InvalidJSON         = 9000
	TooManyLocations    = 9001
	InternalError       = 9999
)

type ErrorInfo struct {
	Message string
}

var weatherAPIErrors = map[int]ErrorInfo{
	KeyNotProvided:      {Message: "Authentication configuration error"},
	LocationNotProvided: {Message: "Location parameter required"},
	InvalidURL:          {Message: "Invalid request format"},
	LocationNotFound:    {Message: "City not found"},
	InvalidKey:          {Message: "Authentication error"},
	QuotaExceeded:       {Message: "Service temporarily unavailable - quota exceeded"},
	KeyDisabled:         {Message: "Service temporarily unavailable"},
	NoAccess:            {Message: "Service feature not available"},
	InvalidJSON:         {Message: "Invalid request data format"},
	TooManyLocations:    {Message: "Too many locations in request"},
	InternalError:       {Message: "Weather service temporarily unavailable"},
}

func (c *Client) mapError(code int, message string) error {
	if code == 0 {
		return nil
	}

	if errorInfo, exists := weatherAPIErrors[code]; exists {
		log.Printf("WeatherAPI error - Code: %d, Original: %s", code, message)
		return weather.NewProviderError(c.Name(), code, errorInfo.Message)
	}

	log.Printf("Unknown WeatherAPI error - Code: %d, Message: %s", code, message)
	return weather.NewProviderError(c.Name(), code, "Weather service temporarily unavailable")
}
