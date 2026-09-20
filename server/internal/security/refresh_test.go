package security_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

func TestRefreshTokensAreRandomAndHashStable(t *testing.T) {
	first, err := security.NewRefreshToken()
	require.NoError(t, err)
	second, err := security.NewRefreshToken()
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	require.Len(t, security.TokenHash(first), 64)
	require.Equal(t, security.TokenHash(first), security.TokenHash(first))
}
