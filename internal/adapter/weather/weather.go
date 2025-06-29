package weather

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"weather-api/internal/core/domain"
	"weather-api/internal/util/jsonutil"
)

type Provider interface {
	GetWeather(ctx context.Context, city string) (domain.Weather, error)
	CheckCityExists(ctx context.Context, city string) error
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient HTTPDoer
}

func NewWeatherAPIClient(apiKey, baseURL string,
	httpClient HTTPDoer,
) *APIClient {
	return &APIClient{
		apiKey: apiKey, baseURL: baseURL,
		httpClient: httpClient,
	}
}

func apiToDomain(w Response) domain.Weather {
	return domain.Weather{
		Temperature: w.TempC,
		Humidity:    w.Humidity,
		Description: w.Condition.Text,
	}
}

func (c *APIClient) GetWeather(ctx context.Context, city string) (domain.Weather, error) {
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

	env, err := jsonutil.Decode[currentEnvelope](resp.Body)
	if err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		return domain.Weather{}, newAPIError(500, err.Error())
	}

	if err := mapAPIError(env.Error); err != nil {
		log.Printf("API Error detected: %v", err)
		return domain.Weather{}, err
	}

	return apiToDomain(env.Current), nil
}

func (c *APIClient) CheckCityExists(ctx context.Context, city string) error {
	log.Printf("Checking if city exists: %s", city)

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/search.json?key="+c.apiKey+"&q="+url.QueryEscape(city),
		nil,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			log.Printf("Failed to close response body: %v", cerr)
			err = cerr
		}
	}()

	var results []SearchItem
	results, err = jsonutil.Decode[[]SearchItem](resp.Body)
	if err != nil {
		log.Printf("Failed to decode results: %v", err)
		return err
	}

	if len(results) == 0 {
		log.Printf("City %s not found in APIClient database", city)
		return domain.ErrCityNotFound
	}

	log.Printf("City %s exists in APIClient database", city)
	return nil
}
