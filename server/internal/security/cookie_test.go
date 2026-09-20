package security_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

func TestRefreshCookieSecurityAttributes(t *testing.T) {
	manager := security.CookieManager{Name: "refresh", Secure: true, TTL: 24 * time.Hour}
	header := manager.Set("opaque-token", time.Unix(100, 0))
	for _, expected := range []string{"HttpOnly", "Secure", "SameSite=Lax", "Path=/api/v1/auth"} {
		require.True(t, strings.Contains(header, expected), header)
	}
	require.NotContains(t, manager.Clear(), "opaque-token")
}
