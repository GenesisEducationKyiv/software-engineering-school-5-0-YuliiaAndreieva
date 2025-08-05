package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SMTP     SMTPConfig
	Server   ServerConfig
	Timeout  TimeoutConfig
	RabbitMQ RabbitMQConfig
	Logging  LoggingConfig
}

type SMTPConfig struct {
	Host string `envconfig:"SMTP_HOST" default:"smtp.gmail.com"`
	Port int    `envconfig:"SMTP_PORT" default:"587"`
	User string `envconfig:"SMTP_USER" required:"true"`
	Pass string `envconfig:"SMTP_PASS" required:"true"`
}

type ServerConfig struct {
	Port                   string `envconfig:"SERVER_PORT" required:"true"`
	GRPCPort               string `envconfig:"GRPC_PORT" required:"true"`
	BaseURL                string `envconfig:"BASE_URL" required:"true"`
	SubscriptionServiceURL string `envconfig:"SUBSCRIPTION_SERVICE_URL" required:"true"`
}

type TimeoutConfig struct {
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
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
