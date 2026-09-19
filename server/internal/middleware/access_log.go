package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func AccessLog(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			started := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			logger.Info("http_request",
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"bytes_out", c.Response().Size,
				"duration_ms", time.Since(started).Milliseconds(),
			)
			return nil
		}
	}
}
