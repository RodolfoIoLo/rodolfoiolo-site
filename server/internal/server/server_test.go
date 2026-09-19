package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/server"
	"github.com/stretchr/testify/require"
)

type healthyDatabase struct{}

func (healthyDatabase) Ping(context.Context) error { return nil }

func TestServerMiddlewareAndHealthRoute(t *testing.T) {
	application := server.New(server.Dependencies{
		Database: healthyDatabase{},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	application.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.NotEmpty(t, recorder.Header().Get("X-Request-Id"))
}
