package handler_test

import (
	"context"
	"testing"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/service"
	"github.com/stretchr/testify/require"
)

type authUseCases struct {
	loginResult service.AuthResult
	loginError  error
}

func (f authUseCases) Login(context.Context, string, string, string, string) (service.AuthResult, error) {
	return f.loginResult, f.loginError
}
func (f authUseCases) Refresh(context.Context, string, string, string) (service.AuthResult, error) {
	return service.AuthResult{}, domain.ErrUnauthorized
}
func (f authUseCases) Logout(context.Context, string) error { return nil }
func (f authUseCases) CurrentUser(context.Context, int64) (repository.User, error) {
	return repository.User{}, domain.ErrNotFound
}

func TestLoginMapsSuccessfulSession(t *testing.T) {
	h := handler.NewAuth(authUseCases{loginResult: service.AuthResult{
		AccessToken: "access", ExpiresIn: 900, RefreshToken: "refresh",
	}}, security.CookieManager{Name: "refresh", TTL: 24 * time.Hour})

	response, err := h.Login(context.Background(), oapi.LoginRequestObject{
		Body: &oapi.LoginRequest{Username: "owner", Password: "correct horse battery staple"},
	})
	require.NoError(t, err)
	success, ok := response.(oapi.Login200JSONResponse)
	require.True(t, ok)
	require.Equal(t, "access", success.Body.Data.AccessToken)
	require.Contains(t, success.Headers.SetCookie, "HttpOnly")
}

func TestLoginDoesNotRevealCredentialFailure(t *testing.T) {
	h := handler.NewAuth(authUseCases{loginError: domain.ErrUnauthorized}, security.CookieManager{Name: "refresh"})
	response, err := h.Login(context.Background(), oapi.LoginRequestObject{
		Body: &oapi.LoginRequest{Username: "unknown", Password: "wrong-password"},
	})
	require.NoError(t, err)
	_, ok := response.(oapi.Login401JSONResponse)
	require.True(t, ok)
}
