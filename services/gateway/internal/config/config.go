package config

import (
	"os"
)

type Config struct {
	Port                   string
	WeatherServiceURL      string
	SubscriptionServiceURL string
}

func LoadConfig() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8080"),
		WeatherServiceURL:      getEnv("WEATHER_SERVICE_URL", "http://localhost:8081"),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
