package security_test

import (
	"testing"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/stretchr/testify/require"
)

func TestVisitorSignatureRejectsTampering(t *testing.T) {
	signer, err := security.NewVisitorSigner([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	id, signed, err := signer.New()
	require.NoError(t, err)
	verified, ok := signer.Verify(signed)
	require.True(t, ok)
	require.Equal(t, id, verified)
	_, ok = signer.Verify(signed + "x")
	require.False(t, ok)
}
