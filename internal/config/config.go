// Package config reads application settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Environment variable names
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

const defaultJWTSecret = "dev-secret-change-me"

// Config holds all settings the application needs to run.
type Config struct {
	AppEnv    string
	Port      string
	DBPath    string
	JWTSecret string
}

// Load returns the configuration.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read .env file: %w", err)
	}

	cfg := Config{
		AppEnv:    strings.ToLower(getEnv("APP_ENV", EnvDevelopment)),
		Port:      getEnv("PORT", "3000"),
		DBPath:    getEnv("DB_PATH", "siri_icecream_shop.db"),
		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-in-production"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c Config) validate() error {
	var problems []string

	if c.AppEnv != EnvDevelopment && c.AppEnv != EnvProduction {
		problems = append(problems, `APP_ENV must be "development" or "production"`)
	}
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		problems = append(problems, "PORT must be a number between 1 and 65535")
	}
	if strings.TrimSpace(c.DBPath) == "" {
		problems = append(problems, "DB_PATH must not be empty")
	}
	if c.IsProduction() {
		if c.JWTSecret == defaultJWTSecret {
			problems = append(problems, "JWT_SECRET must be set in production")
		}
		if len(c.JWTSecret) < 32 {
			problems = append(problems, "JWT_SECRET must be at least 32 characters in production")
		}
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

// IsProduction reports whether the app runs in production.
func (c Config) IsProduction() bool {
	return c.AppEnv == EnvProduction
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
