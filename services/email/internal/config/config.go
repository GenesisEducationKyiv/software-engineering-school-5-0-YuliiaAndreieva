package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTP   SMTPConfig
	Server ServerConfig
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

type ServerConfig struct {
	Port    string
	BaseURL string
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
			Port:    getEnv("SERVER_PORT", "8081"),
			BaseURL: getEnv("BASE_URL", "http://localhost:8081"),
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