package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/requestctx"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

type verifierStub struct {
	identity security.Identity
	err      error
}

func (v verifierStub) Verify(string) (security.Identity, error) { return v.identity, v.err }

type capturedContext struct {
	metadata requestctx.Metadata
	identity requestctx.Identity
	hasID    bool
	refresh  string
}

func runAuthContext(t *testing.T, verifier AccessVerifier, request *http.Request) capturedContext {
	t.Helper()
	e := echo.New()
	recorder := httptest.NewRecorder()

	var captured capturedContext
	next := func(c echo.Context) error {
		ctx := c.Request().Context()
		captured.metadata = requestctx.MetadataFrom(ctx)
		captured.identity, captured.hasID = requestctx.IdentityFrom(ctx)
		captured.refresh = requestctx.RefreshTokenFrom(ctx)
		return nil
	}
	require.NoError(t, AuthContext(verifier, "refresh_token", []byte("visitor-secret"))(next)(e.NewContext(request, recorder)))
	return captured
}

func TestTruncateKeepsShortValuesIntact(t *testing.T) {
	require.Equal(t, "curl/8.0", truncate("curl/8.0", 500))
}

// User-Agent 是攻击者可控的输入。裸切片 value[:500] 会落在多字节字符中间，
// 产生的非法字节序列会被 PostgreSQL 以 22021 拒绝，把登录变成 500。
func TestTruncateNeverSplitsAMultibyteRune(t *testing.T) {
	value := strings.Repeat("浏览器", 200)
	require.Greater(t, len(value), 500)
	require.False(t, utf8.ValidString(value[:500]), "前提：裸切片确实切在字符中间")

	trimmed := truncate(value, 500)

	require.LessOrEqual(t, len(trimmed), 500)
	require.True(t, utf8.ValidString(trimmed), "截断结果必须是合法 UTF-8")
	require.True(t, strings.HasPrefix(value, trimmed), "只能是截断，不能改写内容")
}

func TestTruncateStripsInvalidBytes(t *testing.T) {
	require.Equal(t, "abc", truncate("a\xff\xfebc", 500))
}

func TestBearerToken(t *testing.T) {
	for name, testCase := range map[string]struct {
		header   string
		expected string
	}{
		"standard":         {"Bearer signed-access", "signed-access"},
		"case insensitive": {"bearer signed-access", "signed-access"},
		"wrong scheme":     {"Basic signed-access", ""},
		"missing token":    {"Bearer", ""},
		"too many parts":   {"Bearer a b", ""},
		"empty":            {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, testCase.expected, bearerToken(testCase.header))
		})
	}
}

func TestAuthContextPopulatesMetadataAndRefreshToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	request.Header.Set("User-Agent", "Mozilla/5.0")
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-refresh"})

	captured := runAuthContext(t, verifierStub{err: errors.New("no token")}, request)

	require.Equal(t, "Mozilla/5.0", captured.metadata.UserAgent)
	require.NotEmpty(t, captured.metadata.IPHash)
	require.NotContains(t, captured.metadata.IPHash, "192.0.2.1", "IP 只能以 HMAC 形式出现，不得存原文")
	require.Equal(t, "raw-refresh", captured.refresh)
	require.False(t, captured.hasID)
}

// Cookie 名同时出现在配置、CookieManager 和契约三处。
// 读错名字的症状是「浏览器里 Cookie 明明在，refresh 却一直 401」。
func TestAuthContextOnlyReadsTheConfiguredCookieName(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "personal_site_refresh", Value: "raw-refresh"})

	captured := runAuthContext(t, verifierStub{}, request)

	require.Empty(t, captured.refresh)
}

func TestAuthContextExtractsIdentityFromValidBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer signed-access")

	captured := runAuthContext(t, verifierStub{identity: security.Identity{UserID: 7, Role: "admin"}}, request)

	require.True(t, captured.hasID)
	require.Equal(t, int64(7), captured.identity.UserID)
	require.Equal(t, "admin", captured.identity.Role)
}

func TestAuthContextIgnoresRejectedBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer tampered")

	captured := runAuthContext(t, verifierStub{err: errors.New("bad signature")}, request)

	require.False(t, captured.hasID, "无效令牌不能留下身份")
}
