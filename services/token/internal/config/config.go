package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	JWT     JWTConfig
	Timeout TimeoutConfig
	Logging LoggingConfig
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

type LoggingConfig struct {
	Initial    int
	Thereafter int
	Tick       time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET"),
		},
		Timeout: TimeoutConfig{
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 5*time.Second),
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
