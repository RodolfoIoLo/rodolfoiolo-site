package middleware

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

func CSRF(allowedOrigins []string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodGet || c.Request().Method == http.MethodHead || c.Request().Method == http.MethodOptions {
				return next(c)
			}
			origin := c.Request().Header.Get(echo.HeaderOrigin)
			if _, ok := allowed[origin]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, "origin is not allowed")
			}
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return echo.NewHTTPError(http.StatusForbidden, "origin is invalid")
			}
			if c.Request().Header.Get("X-Requested-With") != "XMLHttpRequest" {
				return echo.NewHTTPError(http.StatusForbidden, "request marker is required")
			}
			return next(c)
		}
	}
}
