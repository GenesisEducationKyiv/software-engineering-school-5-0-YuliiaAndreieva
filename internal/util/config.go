package util

import (
	"os"
	"strconv"
)

type Config struct {
	DBConnStr     string
	WeatherAPIKey string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	Port          int
	BaseUrl       string
}

func LoadConfig() (*Config, error) {
	return &Config{
		DBConnStr:     os.Getenv("DB_CONN_STR"),
		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
		BaseUrl:       GetEnv("BASE_URL", "http://localhost:8080"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      GetEnv("SMTP_PORT", 587),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPass:      os.Getenv("SMTP_PASS"),
		Port:          GetEnv("PORT", 8080),
	}, nil
}

func GetEnv[T any](key string, defaultValue T) T {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return defaultValue
	}

	switch any(defaultValue).(type) {
	case int:
		if i, err := strconv.Atoi(val); err == nil {
			return any(i).(T)
		}
	case string:
		return any(val).(T)
	}
	return defaultValue
}

func GetBaseURL() string {
	return GetEnv("BASE_URL", "http://localhost:8080")
}
