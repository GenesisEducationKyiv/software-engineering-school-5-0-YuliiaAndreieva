package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port                   string `envconfig:"PORT" required:"true"`
	WeatherServiceURL      string `envconfig:"WEATHER_SERVICE_URL" required:"true"`
	SubscriptionServiceURL string `envconfig:"SUBSCRIPTION_SERVICE_URL" required:"true"`
	Timeout                TimeoutConfig
	Logging                LoggingConfig
}

type TimeoutConfig struct {
	HTTPClientTimeout time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"30s"`
	ShutdownTimeout   time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
}

type LoggingConfig struct {
	Initial    int           `envconfig:"LOG_INITIAL" default:"100"`
	Thereafter int           `envconfig:"LOG_THEREAFTER" default:"100"`
	Tick       time.Duration `envconfig:"LOG_TICK" default:"1s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
