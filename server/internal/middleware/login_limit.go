package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type LoginLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorLimiter
	now      func() time.Time
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{visitors: make(map[string]*visitorLimiter), now: time.Now}
}

func (l *LoginLimiter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method != http.MethodPost || c.Path() != "/api/v1/auth/login" {
			return next(c)
		}
		key := c.RealIP()
		l.mu.Lock()
		entry, ok := l.visitors[key]
		if !ok {
			entry = &visitorLimiter{limiter: rate.NewLimiter(rate.Every(12*time.Second), 5)}
			l.visitors[key] = entry
		}
		entry.lastSeen = l.now()
		allowed := entry.limiter.Allow()
		if len(l.visitors) > 1000 {
			cutoff := l.now().Add(-time.Hour)
			for visitor, candidate := range l.visitors {
				if candidate.lastSeen.Before(cutoff) {
					delete(l.visitors, visitor)
				}
			}
		}
		l.mu.Unlock()
		if !allowed {
			return echo.NewHTTPError(http.StatusTooManyRequests, "too many login attempts")
		}
		return next(c)
	}
}
