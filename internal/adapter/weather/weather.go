package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"weather-api/internal/core/domain"
)

type WeatherService struct {
	apiKey string
}

func NewWeatherService(apiKey string) *WeatherService {
	return &WeatherService{apiKey: apiKey}
}

func (w *WeatherService) GetWeather(city string) (domain.Weather, error) {
	url := "http://api.weatherapi.com/v1/current.json?key=" + w.apiKey + "&q=" + city
	resp, err := http.Get(url)
	if err != nil {
		return domain.Weather{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Weather{}, err
	}

	var data struct {
		Current struct {
			TempC     float64 `json:"temp_c"`
			Humidity  int     `json:"humidity"`
			Condition struct {
				Text string `json:"text"`
			} `json:"condition"`
		} `json:"current"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return domain.Weather{}, err
	}

	if data.Error.Code != 0 {
		if data.Error.Code == 1006 {
			return domain.Weather{}, domain.ErrCityNotFound
		}
		return domain.Weather{}, fmt.Errorf("API error: %s", data.Error.Message)
	}

	if data.Current.TempC == 0 && data.Current.Humidity == 0 && data.Current.Condition.Text == "" {
		return domain.Weather{}, domain.ErrCityNotFound
	}

	return domain.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}
