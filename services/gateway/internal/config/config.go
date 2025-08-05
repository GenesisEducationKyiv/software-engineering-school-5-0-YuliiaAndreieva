package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   string
	WeatherServiceURL      string
	SubscriptionServiceURL string
	Timeout                TimeoutConfig
	Logging                LoggingConfig
}

type TimeoutConfig struct {
	HTTPClientTimeout time.Duration
	ShutdownTimeout   time.Duration
}

type LoggingConfig struct {
	Initial    int
	Thereafter int
	Tick       time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:                   getEnv("PORT"),
		WeatherServiceURL:      getEnv("WEATHER_SERVICE_URL"),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL"),
		Timeout: TimeoutConfig{
			HTTPClientTimeout: getDurationEnv("HTTP_CLIENT_TIMEOUT", 30*time.Second),
			ShutdownTimeout:   getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		Logging: LoggingConfig{
			Initial:    getIntEnv("LOG_INITIAL", 100),
			Thereafter: getIntEnv("LOG_THEREAFTER", 100),
			Tick:       getDurationEnv("LOG_TICK", 1*time.Second),
		},
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return "" // Changed from defaultValue to "" to indicate it's required
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
