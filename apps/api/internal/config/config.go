// Package config handles application configuration loaded from environment
// variables using the env library. All values have sensible defaults for local
// development; production deploys override them via the environment.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/execrc/betteroute/internal/config.Version=1.0.0"
var Version = "dev"

// Config holds application-level settings parsed from environment variables.
type Config struct {
	Env  string `env:"APP_ENV" envDefault:"development"`
	Port int    `env:"PORT"    envDefault:"8080"`
}

// Load parses environment variables into Config and validates the result.
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parsing env: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}
	return &cfg, nil
}

// validate enforces business rules that struct tags cannot express.
func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be 1–65535, got %d", c.Port)
	}

	switch c.Env {
	case "development", "staging", "production":
		// ok
	default:
		return fmt.Errorf("APP_ENV must be development, staging, or production, got %q", c.Env)
	}

	return nil
}

// IsDevelopment reports whether the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
