package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	DatabaseURL   string
	HTTPAddr      string
	HealthTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		HTTPAddr:      ":8080",
		HealthTimeout: 2 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = os.Getenv("DEV_DATABASE_URL")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL or DEV_DATABASE_URL is required")
	}

	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if timeout := os.Getenv("HEALTH_TIMEOUT"); timeout != "" {
		parsed, err := time.ParseDuration(timeout)
		if err != nil {
			return Config{}, err
		}
		cfg.HealthTimeout = parsed
	}

	return cfg, nil
}
