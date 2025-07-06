package configutil

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBConnStr             string        `envconfig:"DB_CONN_STR" required:"true"`
	WeatherAPIKey         string        `envconfig:"WEATHER_API_KEY" required:"true"`
	WeatherAPIBaseURL     string        `envconfig:"WEATHER_API_BASE_URL" default:"http://api.weatherapi.com/v1"`
	OpenWeatherMapAPIKey  string        `envconfig:"OPENWEATHERMAP_API_KEY" required:"true"`
	OpenWeatherMapBaseURL string        `envconfig:"OPENWEATHERMAP_BASE_URL" default:"https://api.openweathermap.org/data/2.5"`
	SMTPHost              string        `envconfig:"SMTP_HOST" required:"true"`
	SMTPPort              int           `envconfig:"SMTP_PORT" default:"587"`
	SMTPUser              string        `envconfig:"SMTP_USER" required:"true"`
	SMTPPass              string        `envconfig:"SMTP_PASS" required:"true"`
	Port                  int           `envconfig:"PORT" default:"8080"`
	BaseURL               string        `envconfig:"BASE_URL" default:"http://localhost:8080"`
	HTTPReadTimeout       time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	HTTPWriteTimeout      time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	RedisAddress          string        `envconfig:"REDIS_ADDRESS" default:"localhost:6379"`
	RedisTTL              time.Duration `envconfig:"REDIS_TTL" default:"10m"`
	RedisDialTimeout      time.Duration `envconfig:"REDIS_DIAL_TIMEOUT" default:"5s"`
	RedisReadTimeout      time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"3s"`
	RedisWriteTimeout     time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"3s"`
	RedisPoolSize         int           `envconfig:"REDIS_POOL_SIZE" default:"10"`
	RedisMinIdleConns     int           `envconfig:"REDIS_MIN_IDLE_CONNS" default:"5"`
	HTTPClientTimeout     time.Duration `envconfig:"HTTP_CLIENT_TIMEOUT" default:"5s"`
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
