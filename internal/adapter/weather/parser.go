package weather

import (
	"encoding/json"
	"io"
	"weather-api/internal/core/domain"
)

type WeatherParser interface {
	ParseResponse(r io.Reader) (weatherResponse, error)
	MapToDomain(data weatherResponse) domain.Weather
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseResponse(r io.Reader) (weatherResponse, error) {
	var data weatherResponse
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return weatherResponse{}, err
	}
	return data, nil
}

func (p *Parser) MapToDomain(data weatherResponse) domain.Weather {
	return domain.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}
}
