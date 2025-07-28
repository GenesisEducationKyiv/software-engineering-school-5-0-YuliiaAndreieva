package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Email    EmailConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type EmailConfig struct {
	ServiceURL string
}

type DatabaseConfig struct {
	DSN string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8082"),
		},
		Email: EmailConfig{
			ServiceURL: getEnv("EMAIL_SERVICE_URL", "http://localhost:8081"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "host=localhost user=postgres password=postgres dbname=subscriptions port=5432 sslmode=disable"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
