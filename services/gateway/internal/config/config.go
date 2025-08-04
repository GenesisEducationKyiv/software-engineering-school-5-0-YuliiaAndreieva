package config

import (
	"os"
	"time"
)

type Config struct {
	Port                   string
	WeatherServiceURL      string
	SubscriptionServiceURL string
	Timeout                TimeoutConfig
}

type TimeoutConfig struct {
	HTTPClientTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8080"),
		WeatherServiceURL:      getEnv("WEATHER_SERVICE_URL", "http://localhost:8081"),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8082"),
		Timeout: TimeoutConfig{
			HTTPClientTimeout: getDurationEnv("HTTP_CLIENT_TIMEOUT", 30*time.Second),
			ShutdownTimeout:   getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
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
