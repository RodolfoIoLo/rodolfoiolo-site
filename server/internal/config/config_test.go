package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/config"
)

func TestLoadUsesSafeDevelopmentDefaults(t *testing.T) {
	for _, key := range []string{
		"APP_ENV", "HTTP_ADDRESS", "DATABASE_URL", "DATABASE_MAX_CONNS",
		"DATABASE_MIN_CONNS", "DATABASE_TIMEOUT", "PUBLIC_SITE_URL", "ALLOWED_ORIGINS",
		"REFRESH_COOKIE_NAME",
	} {
		t.Setenv(key, "")
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "development", cfg.Environment)
	require.Equal(t, int32(10), cfg.DatabaseMaxConns)
	require.Equal(t, 5*time.Second, cfg.DatabaseTimeout)
	require.Equal(t, "refresh_token", cfg.RefreshCookieName)
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

func TestLoadRejectsInvalidRefreshCookieName(t *testing.T) {
	t.Setenv("REFRESH_COOKIE_NAME", "refresh token")

	_, err := config.Load()
	require.ErrorContains(t, err, "REFRESH_COOKIE_NAME")
}

// Cookie 名同时出现在「服务端写/读」和「API 契约声明」两侧，
// 两侧不一致的症状是「浏览器里 Cookie 明明在，refresh 却一直 401」。
// 这个测试把配置默认值和生成的契约绑在一起，改一侧不改另一侧就会红。
func TestRefreshCookieNameMatchesContract(t *testing.T) {
	t.Setenv("REFRESH_COOKIE_NAME", "")

	cfg, err := config.Load()
	require.NoError(t, err)

	swagger, err := oapi.GetSwagger()
	require.NoError(t, err)

	scheme, ok := swagger.Components.SecuritySchemes["RefreshCookie"]
	require.True(t, ok, "securitySchemes.RefreshCookie 在契约里不存在")

	require.Equal(t, scheme.Value.Name, cfg.RefreshCookieName,
		"REFRESH_COOKIE_NAME 的默认值必须等于 securitySchemes.RefreshCookie.name")
}
