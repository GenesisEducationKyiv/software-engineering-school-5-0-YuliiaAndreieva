package configutil

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBConnStr     string `envconfig:"DB_CONN_STR" required:"true"`
	WeatherAPIKey string `envconfig:"WEATHER_API_KEY" required:"true"`
	SMTPHost      string `envconfig:"SMTP_HOST" required:"true"`
	SMTPPort      int    `envconfig:"SMTP_PORT" default:"587"`
	SMTPUser      string `envconfig:"SMTP_USER" required:"true"`
	SMTPPass      string `envconfig:"SMTP_PASS" required:"true"`
	Port          int    `envconfig:"PORT" default:"8080"`
	BaseURL       string `envconfig:"BASE_URL" default:"http://localhost:8080"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetBaseURL() string {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return "http://localhost:8080"
	}
	return cfg.BaseURL
}
