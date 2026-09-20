package security_test

import (
	"testing"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/stretchr/testify/require"
)

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := security.HashPassword("correct horse battery staple")
	require.NoError(t, err)
	require.NotContains(t, hash, "correct horse")
	require.True(t, security.VerifyPassword(hash, "correct horse battery staple"))
	require.False(t, security.VerifyPassword(hash, "wrong password"))
}

func TestPasswordRejectsShortInput(t *testing.T) {
	_, err := security.HashPassword("too-short")
	require.Error(t, err)
}
