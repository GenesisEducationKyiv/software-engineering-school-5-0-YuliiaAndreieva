package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SubscriptionServiceURL string        `envconfig:"SUBSCRIPTION_SERVICE_URL" required:"true"`
	WeatherServiceURL      string        `envconfig:"WEATHER_SERVICE_URL" required:"true"`
	EmailServiceURL        string        `envconfig:"EMAIL_SERVICE_URL" required:"true"`
	SubscriptionGRPCURL    string        `envconfig:"SUBSCRIPTION_GRPC_URL" default:"subscription-service:9090"`
	EmailGRPCURL           string        `envconfig:"EMAIL_GRPC_URL" default:"email-service:9091"`
	WeatherGRPCURL         string        `envconfig:"WEATHER_GRPC_URL" default:"weather-service:9092"`
	Port                   int           `envconfig:"PORT" default:"8085"`
	WorkerAmount           int           `envconfig:"WORKER_AMOUNT" default:"10"`
	PageSize               int           `envconfig:"PAGE_SIZE" default:"100"`
	HTTPClientTimeout      time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"10s"`
	HTTPReadTimeout        time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	HTTPWriteTimeout       time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	ShutdownTimeout        time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
	LogInitial             int           `envconfig:"LOG_INITIAL" default:"100"`
	LogThereafter          int           `envconfig:"LOG_THEREAFTER" default:"100"`
	LogTick                time.Duration `envconfig:"LOG_TICK" default:"1s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
