package middleware

import (
	"net/http"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/requestctx"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/labstack/echo/v4"
)

func VisitorIdentity(name string, secure bool, signer security.VisitorSigner) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var id string
			if cookie, err := c.Cookie(name); err == nil {
				id, _ = signer.Verify(cookie.Value)
			}
			if id == "" {
				var signed string
				var err error
				id, signed, err = signer.New()
				if err != nil {
					return err
				}
				c.SetCookie(&http.Cookie{Name: name, Value: signed, Path: "/api/v1", MaxAge: 31536000, Expires: time.Now().UTC().Add(365 * 24 * time.Hour), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
			}
			ctx := c.Request().Context()
			metadata := requestctx.MetadataFrom(ctx)
			metadata.VisitorID = id
			c.SetRequest(c.Request().WithContext(requestctx.WithMetadata(ctx, metadata)))
			return next(c)
		}
	}
}
