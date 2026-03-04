package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the API service.
type Config struct {
	DatabaseURL      string
	JWTSecret        string
	JWTRefreshSecret string
	Port             string
	Env              string
	AllowedOrigins   []string
}

// Load reads configuration from environment variables and returns a Config.
func Load() Config {
	return Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/skillisland"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		AllowedOrigins:   strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
