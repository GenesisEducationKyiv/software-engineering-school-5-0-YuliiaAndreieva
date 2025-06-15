package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"weather-api/internal/core/domain"
)

type Provider interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
	ValidateCity(ctx context.Context, city string) error
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type WeatherValidator interface {
	ValidateResponse(data WeatherResponse) error
}

type WeatherAPIClient struct {
	apiKey     string
	baseURL    string
	httpClient HTTPDoer
	parser     WeatherParser
	validator  WeatherValidator
}

func NewWeatherAPIClient(apiKey, baseURL string,
	httpClient HTTPDoer,
	parser WeatherParser,
	validator WeatherValidator,
) *WeatherAPIClient {
	return &WeatherAPIClient{
		apiKey: apiKey, baseURL: baseURL,
		httpClient: httpClient,
		parser:     parser,
		validator:  validator,
	}
}

func (c *WeatherAPIClient) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/current.json?key="+c.apiKey+"&q="+url.QueryEscape(city),
		nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.Weather{}, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	data, err := c.parser.ParseResponse(resp.Body)
	if err != nil {
		return domain.Weather{}, err
	}
	if err = c.validator.ValidateResponse(data); err != nil {
		return domain.Weather{}, err
	}
	return c.parser.MapToDomain(data), nil
}

func (c *WeatherAPIClient) ValidateCity(ctx context.Context, city string) error {
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/search.json?key="+c.apiKey+"&q="+url.QueryEscape(city),
		nil,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var results []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return err
	}

	if len(results) == 0 {
		return domain.ErrCityNotFound
	}
	return nil
}
