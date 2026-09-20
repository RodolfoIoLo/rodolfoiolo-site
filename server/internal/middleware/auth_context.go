package middleware

import (
	"strings"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/requestctx"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/labstack/echo/v4"
)

type AccessVerifier interface {
	Verify(string) (security.Identity, error)
}

func AuthContext(verifier AccessVerifier, refreshCookieName string, visitorSecret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			metadata := requestctx.Metadata{
				UserAgent: truncate(c.Request().UserAgent(), 500),
				IPHash:    security.HMACHash(visitorSecret, c.RealIP()),
			}
			ctx = requestctx.WithMetadata(ctx, metadata)

			if cookie, err := c.Cookie(refreshCookieName); err == nil {
				ctx = requestctx.WithRefreshToken(ctx, cookie.Value)
			}
			if raw := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization)); raw != "" {
				if identity, err := verifier.Verify(raw); err == nil {
					ctx = requestctx.WithIdentity(ctx, requestctx.Identity{UserID: identity.UserID, Role: identity.Role})
				}
			}
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

// truncate 把字符串限制在 maximum 字节以内，并保证结果是合法 UTF-8。
//
// 直接写 value[:maximum] 会把多字节字符切成两半，产生非法字节序列。
// 这个值随后写入 refresh_tokens.user_agent（VARCHAR(500)），
// PostgreSQL 会以 SQLSTATE 22021 invalid byte sequence 拒绝整条 INSERT ——
// 症状是登录返回 500 而不是 401，而且只在 User-Agent 含非 ASCII 字符时出现。
//
// ToValidUTF8 同时覆盖两种情况：尾部被切断的半个字符，以及上游传来的非法字节。
func truncate(value string, maximum int) string {
	if len(value) > maximum {
		value = value[:maximum]
	}
	return strings.ToValidUTF8(value, "")
}
