package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server  ServerConfig
	JWT     JWTConfig
	Timeout TimeoutConfig
	Logging LoggingConfig
}

type ServerConfig struct {
	Port string `envconfig:"SERVER_PORT" required:"true"`
}

type JWTConfig struct {
	Secret string `envconfig:"JWT_SECRET" required:"true"`
}

type TimeoutConfig struct {
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
}

type LoggingConfig struct {
	Initial    int           `envconfig:"LOG_INITIAL" default:"100"`
	Thereafter int           `envconfig:"LOG_THEREAFTER" default:"100"`
	Tick       time.Duration `envconfig:"LOG_TICK" default:"1s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
