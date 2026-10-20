package server_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/server"
	"github.com/stretchr/testify/require"
)

func TestServerMiddlewareAndLivenessRoute(t *testing.T) {
	signer, err := security.NewVisitorSigner([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	application := server.New(server.Dependencies{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), API: handler.New(handler.Dependencies{}), VisitorSigner: signer, RefreshCookieName: "refresh", VisitorCookieName: "visitor", VisitorSecret: []byte("0123456789abcdef0123456789abcdef"), AllowedOrigins: []string{"http://localhost:3000"}})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	application.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.NotEmpty(t, recorder.Header().Get("X-Request-Id"))
}
