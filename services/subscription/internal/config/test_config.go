package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
)

type TestConfig struct {
	Server   TestServerConfig
	Database TestDatabaseConfig
	RabbitMQ TestRabbitMQConfig
	Email    TestEmailConfig
}

type TestServerConfig struct {
	Port    string `envconfig:"TEST_SERVER_PORT" default:"8082"`
	BaseURL string `envconfig:"TEST_BASE_URL" default:"http://localhost:8082"`
}

type TestDatabaseConfig struct {
	Host     string `envconfig:"TEST_DB_HOST" default:"localhost"`
	Port     string `envconfig:"TEST_DB_PORT" default:"5432"`
	User     string `envconfig:"TEST_DB_USER" default:"test"`
	Password string `envconfig:"TEST_DB_PASSWORD" default:"test"`
	Name     string `envconfig:"TEST_DB_NAME" default:"subscription_test"`
}

type TestRabbitMQConfig struct {
	URL      string `envconfig:"TEST_RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
	Exchange string `envconfig:"TEST_RABBITMQ_EXCHANGE" default:"subscription_events"`
	Queue    string `envconfig:"TEST_RABBITMQ_QUEUE" default:"email_notifications"`
}

type TestEmailConfig struct {
	ServiceURL string `envconfig:"TEST_EMAIL_SERVICE_URL" default:"http://localhost:8081"`
}

func LoadTestConfig() (*TestConfig, error) {
	var cfg TestConfig

	if os.Getenv("TEST_DB_HOST") == "" {
		cfg.Database.Host = "postgres"
	}
	if os.Getenv("TEST_RABBITMQ_URL") == "" {
		cfg.RabbitMQ.URL = "amqp://guest:guest@rabbitmq:5672/"
	}
	if os.Getenv("TEST_BASE_URL") == "" {
		cfg.Server.BaseURL = "http://subscription-service:8082"
	}
	if os.Getenv("TEST_EMAIL_SERVICE_URL") == "" {
		cfg.Email.ServiceURL = "http://fake-email:8081"
	}

	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
