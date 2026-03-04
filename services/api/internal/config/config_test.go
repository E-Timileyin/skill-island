package config_test

import (
	"os"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// Unset env vars to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("ALLOWED_ORIGINS")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Fatalf("expected default env development, got %s", cfg.Env)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Fatalf("expected default allowed origins, got %v", cfg.AllowedOrigins)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://other.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Fatalf("expected env production, got %s", cfg.Env)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Fatalf("expected JWT secret test-secret, got %s", cfg.JWTSecret)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.AllowedOrigins))
	}
}
