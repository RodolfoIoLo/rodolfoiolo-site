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
	Environment, HTTPAddress, DatabaseURL                       string
	DatabaseMaxConns, DatabaseMinConns                          int32
	DatabaseTimeout                                             time.Duration
	PublicSiteURL                                               string
	AllowedOrigins                                              []string
	JWTPrivateKeyFile, JWTPublicKeyFile, JWTIssuer, JWTAudience string
	AccessTTL, RefreshTTL                                       time.Duration
	RefreshCookieName, VisitorCookieName, VisitorSecretFile     string
	CookieSecure                                                bool
	UploadRoot, MediaPublicURL                                  string
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
	accessTTL, err := durationEnvironment("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := durationEnvironment("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	secure, err := boolEnvironment("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Environment: environment("APP_ENV", "development"), HTTPAddress: environment("HTTP_ADDRESS", ":8080"),
		DatabaseURL: environment("DATABASE_URL", "postgres://personal_site:personal_site_dev@localhost:5433/personal_site?sslmode=disable"), DatabaseMaxConns: maxConns, DatabaseMinConns: minConns, DatabaseTimeout: timeout,
		PublicSiteURL: environment("PUBLIC_SITE_URL", "http://localhost:3000"), AllowedOrigins: splitEnvironment("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		JWTPrivateKeyFile: environment("JWT_PRIVATE_KEY_FILE", "../secrets/jwt_private_key.pem"), JWTPublicKeyFile: environment("JWT_PUBLIC_KEY_FILE", "../secrets/jwt_public_key.pem"), JWTIssuer: environment("JWT_ISSUER", "personal-site-api"), JWTAudience: environment("JWT_AUDIENCE", "personal-site-admin"), AccessTTL: accessTTL, RefreshTTL: refreshTTL,
		RefreshCookieName: environment("REFRESH_COOKIE_NAME", "refresh_token"), VisitorCookieName: environment("VISITOR_COOKIE_NAME", "personal_site_visitor"), VisitorSecretFile: environment("VISITOR_SECRET_FILE", "../secrets/visitor_secret.txt"), CookieSecure: secure,
		UploadRoot: environment("UPLOAD_ROOT", "../uploads"), MediaPublicURL: environment("MEDIA_PUBLIC_URL", "http://localhost:8080/uploads"),
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
	if c.DatabaseTimeout <= 0 || c.AccessTTL <= 0 || c.RefreshTTL <= 0 {
		return errors.New("timeouts and token TTLs must be positive")
	}
	for name, rawURL := range map[string]string{"DATABASE_URL": c.DatabaseURL, "PUBLIC_SITE_URL": c.PublicSiteURL, "MEDIA_PUBLIC_URL": c.MediaPublicURL} {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("%s must be an absolute URL", name)
		}
	}
	if len(c.AllowedOrigins) == 0 {
		return errors.New("ALLOWED_ORIGINS must contain at least one origin")
	}
	if c.JWTPrivateKeyFile == "" || c.JWTPublicKeyFile == "" || c.VisitorSecretFile == "" || c.VisitorCookieName == "" {
		return errors.New("authentication file and cookie settings are required")
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
func boolEnvironment(key string, fallback bool) (bool, error) {
	raw := environment(key, "")
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
