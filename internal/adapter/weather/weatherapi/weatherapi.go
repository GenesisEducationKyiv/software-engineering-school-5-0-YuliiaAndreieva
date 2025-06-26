package weatherapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/jsonutil"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient weather.HTTPDoer
	logger     weather.ProviderLogger
}

type ClientOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient weather.HTTPDoer
	Logger     weather.ProviderLogger
}

func NewClient(opts ClientOptions) *Client {
	return &Client{
		apiKey:     opts.APIKey,
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		logger:     opts.Logger,
	}
}

func apiToDomain(w response) domain.Weather {
	return domain.Weather{
		Temperature: w.TempC,
		Humidity:    w.Humidity,
		Description: w.Condition.Text,
	}
}

func (c *Client) Name() string {
	return "WeatherAPI"
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/current.json?key="+c.apiKey+"&q="+url.QueryEscape(city),
		nil)
	if err != nil {
		msg := fmt.Sprintf("unable to create HTTP request: %v", err)
		log.Print(msg)
		return domain.Weather{}, errors.New(msg)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("unable to make HTTP request for city %s: %v", city, err)
		log.Print(msg)
		return domain.Weather{}, errors.New(msg)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	var logBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &logBuffer)

	env, err := jsonutil.Decode[currentEnvelope](teeReader)
	if err != nil {
		msg := fmt.Sprintf("unable to decode JSON: %v", err)
		log.Print(msg)
		return domain.Weather{}, weather.NewProviderError(c.Name(), 500, msg)
	}

	c.logger.Log(c.Name(), logBuffer.Bytes())

	if env.Error.Code != 0 {
		log.Printf("API Error detected: %v", env.Error)
		return domain.Weather{}, c.mapError(env.Error.Code, env.Error.Message)
	}

	return apiToDomain(env.Current), nil
}

func (c *Client) CheckCityExists(ctx context.Context, city string) error {
	log.Printf("Checking if city exists: %s", city)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/search.json?key="+c.apiKey+"&q="+url.QueryEscape(city),
		nil)
	if err != nil {
		msg := fmt.Sprintf("creating request: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("unable to make HTTP request for city %s: %v", city, err)
		log.Print(msg)
		return errors.New(msg)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	var logBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &logBuffer)

	var results []searchItem
	results, err = jsonutil.Decode[[]searchItem](teeReader)
	if err != nil {
		msg := fmt.Sprintf("decoding request: %v", err)
		log.Print(msg)
		return errors.New(msg)
	}

	c.logger.Log(c.Name(), logBuffer.Bytes())

	if len(results) == 0 {
		log.Printf("City %s not found in WeatherAPI database", city)
		return weather.NewProviderError(c.Name(), 404, "city not found")
	}

	log.Printf("City %s exists in WeatherAPI database", city)
	return nil
}
