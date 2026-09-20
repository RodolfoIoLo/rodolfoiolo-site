package security

import (
	"net/http"
	"time"
)

type CookieManager struct {
	Name   string
	Secure bool
	TTL    time.Duration
}

func (m CookieManager) Set(raw string, now time.Time) string {
	return (&http.Cookie{
		Name:     m.Name,
		Value:    raw,
		Path:     "/api/v1/auth",
		Expires:  now.Add(m.TTL),
		MaxAge:   int(m.TTL.Seconds()),
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode,
	}).String()
}

func (m CookieManager) Clear() string {
	return (&http.Cookie{
		Name:     m.Name,
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode,
	}).String()
}
