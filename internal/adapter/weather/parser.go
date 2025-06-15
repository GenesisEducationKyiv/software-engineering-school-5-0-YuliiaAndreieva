package weather

import (
	"encoding/json"
	"io"
	"weather-api/internal/core/domain"
)

type WeatherParser interface {
	ParseResponse(r io.Reader) (WeatherResponse, error)
	MapToDomain(data WeatherResponse) domain.Weather
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseResponse(r io.Reader) (WeatherResponse, error) {
	var data WeatherResponse
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return WeatherResponse{}, err
	}
	return data, nil
}

func (p *Parser) MapToDomain(data WeatherResponse) domain.Weather {
	return domain.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}
}
