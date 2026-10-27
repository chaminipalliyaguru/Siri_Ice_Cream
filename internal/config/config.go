// Package config reads application settings from environment variables.
package config

import "os"

// Config holds all settings the application needs to run.
type Config struct {
	Port      string // HTTP port, e.g. "3000"
	DBPath    string // SQLite database file
	JWTSecret string // secret key used to sign JWT tokens
}

// Load returns the configuration, using sensible defaults when an
// environment variable is not set.
func Load() Config {
	return Config{
		Port:      getEnv("PORT", "3000"),
		DBPath:    getEnv("DB_PATH", "siri_icecream_shop.db"),
		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
