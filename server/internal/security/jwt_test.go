package security_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

func TestJWTIssueAndVerify(t *testing.T) {
	manager, err := security.LoadJWTManager(
		filepath.Join("..", "..", "..", "secrets", "jwt_private_key.pem"),
		filepath.Join("..", "..", "..", "secrets", "jwt_public_key.pem"),
		"personal-site-api", "personal-site-admin", 15*time.Minute,
	)
	require.NoError(t, err)

	raw, expiresAt, err := manager.Issue(security.Identity{UserID: 42, Role: "admin"})
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(15*time.Minute), expiresAt, 5*time.Second)

	identity, err := manager.Verify(raw)
	require.NoError(t, err)
	require.Equal(t, int64(42), identity.UserID)
	require.Equal(t, "admin", identity.Role)
}

func TestJWTRejectsGarbage(t *testing.T) {
	manager, err := security.LoadJWTManager(
		filepath.Join("..", "..", "..", "secrets", "jwt_private_key.pem"),
		filepath.Join("..", "..", "..", "secrets", "jwt_public_key.pem"),
		"personal-site-api", "personal-site-admin", 15*time.Minute,
	)
	require.NoError(t, err)
	_, err = manager.Verify("not-a-jwt")
	require.Error(t, err)
}
