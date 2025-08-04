package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	WeatherAPIKey         string        `envconfig:"WEATHER_API_KEY" required:"true"`
	WeatherAPIBaseURL     string        `envconfig:"WEATHER_API_BASE_URL" default:"http://api.weatherapi.com/v1"`
	OpenWeatherMapAPIKey  string        `envconfig:"OPENWEATHERMAP_API_KEY"`
	OpenWeatherMapBaseURL string        `envconfig:"OPENWEATHERMAP_BASE_URL" default:"https://api.openweathermap.org/data/2.5"`
	Port                  int           `envconfig:"PORT" default:"8084"`
	GRPCPort              int           `envconfig:"GRPC_PORT" default:"9092"`
	BaseURL               string        `envconfig:"BASE_URL" default:"http://localhost:8084"`
	HTTPReadTimeout       time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	HTTPWriteTimeout      time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	RedisAddress          string        `envconfig:"REDIS_ADDRESS" default:"localhost:6379"`
	RedisTTL              time.Duration `envconfig:"REDIS_TTL" default:"30m"`
	RedisDialTimeout      time.Duration `envconfig:"REDIS_DIAL_TIMEOUT" default:"5s"`
	RedisReadTimeout      time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"3s"`
	RedisWriteTimeout     time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"3s"`
	RedisPoolSize         int           `envconfig:"REDIS_POOL_SIZE" default:"10"`
	RedisMinIdleConns     int           `envconfig:"REDIS_MIN_IDLE_CONNS" default:"5"`
	HTTPClientTimeout     time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"10s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetBaseURL() string {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return "http://localhost:8084"
	}
	return cfg.BaseURL
}
