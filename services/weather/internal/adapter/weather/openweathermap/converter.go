package openweathermap

import (
	"weather/internal/core/domain"
)

func convertToDomain(weatherResp *Response) domain.Weather {
	return domain.Weather{
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Description: weatherResp.Weather[0].Description,
	}
}
