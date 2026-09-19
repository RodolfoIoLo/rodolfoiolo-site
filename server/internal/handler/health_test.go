package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
)

type pinger struct {
	err error
}

func (p pinger) Ping(context.Context) error { return p.err }

func TestLivenessDoesNotRequireDatabase(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	h := handler.NewHealth(pinger{err: errors.New("database unavailable")})

	require.NoError(t, h.Liveness(e.NewContext(request, recorder)))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestReadinessReportsConnectedDatabase(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	h := handler.NewHealth(pinger{})

	require.NoError(t, h.Readiness(e.NewContext(request, recorder)))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestReadinessFailsClosed(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	h := handler.NewHealth(pinger{err: errors.New("database unavailable")})

	require.NoError(t, h.Readiness(e.NewContext(request, recorder)))
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "database unavailable")
}
