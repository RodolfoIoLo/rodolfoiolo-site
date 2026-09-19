package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/config"
)

func TestLoadUsesSafeDevelopmentDefaults(t *testing.T) {
	for _, key := range []string{
		"APP_ENV", "HTTP_ADDRESS", "DATABASE_URL", "DATABASE_MAX_CONNS",
		"DATABASE_MIN_CONNS", "DATABASE_TIMEOUT", "PUBLIC_SITE_URL", "ALLOWED_ORIGINS",
	} {
		t.Setenv(key, "")
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "development", cfg.Environment)
	require.Equal(t, int32(10), cfg.DatabaseMaxConns)
	require.Equal(t, 5*time.Second, cfg.DatabaseTimeout)
}

func TestLoadRejectsInvalidConnectionLimits(t *testing.T) {
	t.Setenv("DATABASE_MIN_CONNS", "11")
	t.Setenv("DATABASE_MAX_CONNS", "10")

	_, err := config.Load()
	require.ErrorContains(t, err, "connection limits")
}

func TestLoadRejectsRelativePublicURL(t *testing.T) {
	t.Setenv("PUBLIC_SITE_URL", "/relative")

	_, err := config.Load()
	require.ErrorContains(t, err, "PUBLIC_SITE_URL")
}
