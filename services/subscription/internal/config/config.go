package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Email    EmailConfig
	Token    TokenConfig
	Database DatabaseConfig
	Timeout  TimeoutConfig
	RabbitMQ RabbitMQConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port     string
	GRPCPort string
	BaseURL  string
}

type EmailConfig struct {
	ServiceURL string
}

type TokenConfig struct {
	ServiceURL string
	Expiration string
}

type DatabaseConfig struct {
	DSN string
}

type TimeoutConfig struct {
	HTTPClientTimeout  time.Duration
	ShutdownTimeout    time.Duration
	DatabaseRetryDelay time.Duration
	DatabaseMaxRetries int
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Queue    string
}

type LoggingConfig struct {
	Initial    int
	Thereafter int
	Tick       time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     getEnv("SERVER_PORT", "8082"),
			GRPCPort: getEnv("GRPC_PORT", "9093"),
			BaseURL:  getEnv("BASE_URL", "http://localhost:8082"),
		},
		Email: EmailConfig{
			ServiceURL: getEnv("EMAIL_SERVICE_URL", "http://email-service:8081"),
		},
		Token: TokenConfig{
			ServiceURL: getEnv("TOKEN_SERVICE_URL", "http://token-service:8083"),
			Expiration: getEnv("TOKEN_EXPIRATION", "24h"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "host=postgres user=postgres password=postgres dbname=subscriptions port=5432 sslmode=disable"),
		},
		Timeout: TimeoutConfig{
			HTTPClientTimeout:  getDurationEnv("HTTP_CLIENT_TIMEOUT", 10*time.Second),
			ShutdownTimeout:    getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
			DatabaseRetryDelay: getDurationEnv("DATABASE_RETRY_DELAY", 2*time.Second),
			DatabaseMaxRetries: getIntEnv("DATABASE_MAX_RETRIES", 30),
		},
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", "amqp://admin:password@rabbitmq:5672/"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "subscription_events"),
			Queue:    getEnv("RABBITMQ_QUEUE", "email_notifications"),
		},
		Logging: LoggingConfig{
			Initial:    getIntEnv("LOG_INITIAL", 100),
			Thereafter: getIntEnv("LOG_THEREAFTER", 100),
			Tick:       getDurationEnv("LOG_TICK", 1*time.Second),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
