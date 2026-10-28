package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppEnv != EnvDevelopment || cfg.Port != "3000" || cfg.DBPath != "siri_icecream_shop.db" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
	if cfg.JWTSecret != defaultJWTSecret {
		t.Errorf("default JWT secret is %q, want the defaultJWTSecret constant %q", cfg.JWTSecret, defaultJWTSecret)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantMsg string
	}{
		{"bad port", map[string]string{"PORT": "abc"}, "PORT"},
		{"bad env", map[string]string{"APP_ENV": "staging"}, "APP_ENV"},
		{"default secret in production", map[string]string{"APP_ENV": "production", "JWT_SECRET": ""}, "JWT_SECRET"},
		{"short secret in production", map[string]string{"APP_ENV": "production", "JWT_SECRET": "short"}, "32 characters"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", "")
			t.Setenv("PORT", "")
			t.Setenv("JWT_SECRET", "")
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("got error %v, want it to mention %q", err, tc.wantMsg)
			}
		})
	}
}

func TestLoadProductionValid(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("JWT_SECRET", strings.Repeat("x", 32))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("expected production")
	}
}
