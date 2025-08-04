package config

import (
	"os"
	"time"
)

type Config struct {
	Server  ServerConfig
	JWT     JWTConfig
	Timeout TimeoutConfig
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Secret string
}

type TimeoutConfig struct {
	ShutdownTimeout time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8083"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		Timeout: TimeoutConfig{
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
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
