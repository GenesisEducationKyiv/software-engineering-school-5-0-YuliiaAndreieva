package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTP     SMTPConfig
	Server   ServerConfig
	Timeout  TimeoutConfig
	RabbitMQ RabbitMQConfig
	Logging  LoggingConfig
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

type ServerConfig struct {
	Port     string
	GRPCPort string
	BaseURL  string
}

type TimeoutConfig struct {
	ShutdownTimeout time.Duration
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
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port: getEnvAsInt("SMTP_PORT", 587),
			User: getEnv("SMTP_USER", ""),
			Pass: getEnv("SMTP_PASS", ""),
		},
		Server: ServerConfig{
			Port:     getEnv("SERVER_PORT", "8081"),
			GRPCPort: getEnv("GRPC_PORT", "9091"),
			BaseURL:  getEnv("BASE_URL", "http://localhost:8081"),
		},
		Timeout: TimeoutConfig{
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
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

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
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

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
