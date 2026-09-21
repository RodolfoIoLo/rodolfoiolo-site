package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment       string
	HTTPAddress       string
	DatabaseURL       string
	DatabaseMaxConns  int32
	DatabaseMinConns  int32
	DatabaseTimeout   time.Duration
	PublicSiteURL     string
	AllowedOrigins    []string
	RefreshCookieName string
}

func Load() (Config, error) {
	maxConns, err := int32Environment("DATABASE_MAX_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	minConns, err := int32Environment("DATABASE_MIN_CONNS", 1)
	if err != nil {
		return Config{}, err
	}
	timeout, err := durationEnvironment("DATABASE_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:       environment("APP_ENV", "development"),
		HTTPAddress:       environment("HTTP_ADDRESS", ":8080"),
		DatabaseURL:       environment("DATABASE_URL", "postgres://personal_site:personal_site_dev@localhost:5433/personal_site?sslmode=disable"),
		DatabaseMaxConns:  maxConns,
		DatabaseMinConns:  minConns,
		DatabaseTimeout:   timeout,
		PublicSiteURL:     environment("PUBLIC_SITE_URL", "http://localhost:3000"),
		AllowedOrigins:    splitEnvironment("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		RefreshCookieName: environment("REFRESH_COOKIE_NAME", "refresh_token"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.HTTPAddress == "" {
		return errors.New("HTTP_ADDRESS is required")
	}
	if c.DatabaseMinConns < 0 || c.DatabaseMaxConns < 1 || c.DatabaseMinConns > c.DatabaseMaxConns {
		return errors.New("database connection limits are invalid")
	}
	if c.DatabaseTimeout <= 0 {
		return errors.New("DATABASE_TIMEOUT must be positive")
	}
	for name, rawURL := range map[string]string{
		"DATABASE_URL":    c.DatabaseURL,
		"PUBLIC_SITE_URL": c.PublicSiteURL,
	} {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("%s must be an absolute URL", name)
		}
	}
	if len(c.AllowedOrigins) == 0 {
		return errors.New("ALLOWED_ORIGINS must contain at least one origin")
	}
	if c.RefreshCookieName == "" || strings.ContainsAny(c.RefreshCookieName, " \t;=,") {
		return errors.New("REFRESH_COOKIE_NAME must be a valid cookie name")
	}
	return nil
}

func environment(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return fallback
}

func splitEnvironment(key string, fallback []string) []string {
	raw := environment(key, "")
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func int32Environment(key string, fallback int32) (int32, error) {
	raw := environment(key, "")
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return int32(value), nil
}

func durationEnvironment(key string, fallback time.Duration) (time.Duration, error) {
	raw := environment(key, "")
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return value, nil
}
