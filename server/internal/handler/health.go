package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Pinger interface {
	Ping(context.Context) error
}

type Health struct {
	database Pinger
	now      func() time.Time
}

func NewHealth(database Pinger) *Health {
	return &Health{database: database, now: time.Now}
}

func (h *Health) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":    "ok",
		"timestamp": h.now().UTC().Format(time.RFC3339),
	})
}

func (h *Health) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()
	if err := h.database.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":   "not_ready",
			"database": "disconnected",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status":   "ready",
		"database": "connected",
	})
}
