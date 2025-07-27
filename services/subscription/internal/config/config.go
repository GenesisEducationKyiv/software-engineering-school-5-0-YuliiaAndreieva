package config

import (
	"os"
)

type Config struct {
	Server ServerConfig
	Email  EmailConfig
}

type ServerConfig struct {
	Port string
}

type EmailConfig struct {
	ServiceURL string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8082"),
		},
		Email: EmailConfig{
			ServiceURL: getEnv("EMAIL_SERVICE_URL", "http://localhost:8081"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 