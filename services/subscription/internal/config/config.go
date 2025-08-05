package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server   ServerConfig
	Token    TokenConfig
	Database DatabaseConfig
	Timeout  TimeoutConfig
	RabbitMQ RabbitMQConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port     string `envconfig:"SERVER_PORT" required:"true"`
	GRPCPort string `envconfig:"GRPC_PORT" required:"true"`
	BaseURL  string `envconfig:"BASE_URL" required:"true"`
}

type TokenConfig struct {
	ServiceURL string `envconfig:"TOKEN_SERVICE_URL" required:"true"`
	Expiration string `envconfig:"TOKEN_EXPIRATION" default:"24h"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DATABASE_DSN" default:"host=postgres user=postgres password=postgres dbname=subscriptions port=5432 sslmode=disable"`
}

type TimeoutConfig struct {
	HTTPClientTimeout  time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"10s"`
	ShutdownTimeout    time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
	DatabaseRetryDelay time.Duration `envconfig:"DATABASE_RETRY_DELAY" default:"2s"`
	DatabaseMaxRetries int           `envconfig:"DATABASE_MAX_RETRIES" default:"30"`
}

type RabbitMQConfig struct {
	URL      string `envconfig:"RABBITMQ_URL" default:"amqp://admin:password@rabbitmq:5672/"`
	Exchange string `envconfig:"RABBITMQ_EXCHANGE" default:"subscription_events"`
	Queue    string `envconfig:"RABBITMQ_QUEUE" default:"email_notifications"`
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
