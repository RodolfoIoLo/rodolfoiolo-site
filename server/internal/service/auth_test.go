package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/service"
	"github.com/stretchr/testify/require"
)

type authStore struct {
	user          repository.User
	findUserError error
	createdHash   string
	rotatedHash   string
	revokedHash   string
}

func (s *authStore) FindUserByUsername(context.Context, string) (repository.User, error) {
	return s.user, s.findUserError
}
func (s *authStore) FindUserByID(context.Context, int64) (repository.User, error) { return s.user, nil }
func (s *authStore) CreateRefreshToken(_ context.Context, _ int64, hash string, _ time.Time, _, _ string) error {
	s.createdHash = hash
	return nil
}
func (s *authStore) RotateRefreshToken(_ context.Context, _, next string, _ time.Time, _, _ string) (repository.User, error) {
	s.rotatedHash = next
	return s.user, nil
}
func (s *authStore) RevokeRefreshToken(_ context.Context, hash string) error {
	s.revokedHash = hash
	return nil
}

type issuer struct{}

func (issuer) Issue(identity security.Identity) (string, time.Time, error) {
	if identity.UserID <= 0 {
		return "", time.Time{}, errors.New("bad identity")
	}
	return "signed-access", time.Now().Add(15 * time.Minute), nil
}
func (issuer) TTL() time.Duration { return 15 * time.Minute }

func TestLoginCreatesOnlyHashedRefreshToken(t *testing.T) {
	hash, err := security.HashPassword("correct horse battery staple")
	require.NoError(t, err)
	store := &authStore{user: repository.User{ID: 1, Username: "owner", PasswordHash: hash, Role: "admin"}}
	useCases := service.NewAuthService(store, issuer{}, 30*24*time.Hour)

	result, err := useCases.Login(context.Background(), "owner", "correct horse battery staple", "test", "ip-hash")
	require.NoError(t, err)
	require.Equal(t, "signed-access", result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.Equal(t, security.TokenHash(result.RefreshToken), store.createdHash)
	require.NotEqual(t, result.RefreshToken, store.createdHash)
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	hash, err := security.HashPassword("correct horse battery staple")
	require.NoError(t, err)
	store := &authStore{user: repository.User{ID: 1, PasswordHash: hash, Role: "admin"}}
	useCases := service.NewAuthService(store, issuer{}, 30*24*time.Hour)
	_, err = useCases.Login(context.Background(), "owner", "wrong password", "", "")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

// 数据库故障必须原样上抛，不能被折叠成 ErrUnauthorized。
// 曾经这里顺序写反：err 判断排在密码校验之后，零值 user 的空哈希让
// VerifyPassword 返回 false，于是连接失败也报「用户名或密码错误」。
func TestLoginPropagatesStoreFailure(t *testing.T) {
	storeFailure := errors.New("connection refused")
	store := &authStore{findUserError: storeFailure}
	useCases := service.NewAuthService(store, issuer{}, 30*24*time.Hour)

	_, err := useCases.Login(context.Background(), "owner", "correct horse battery staple", "", "")
	require.ErrorIs(t, err, storeFailure)
	require.NotErrorIs(t, err, domain.ErrUnauthorized)
}

// 用户不存在对外仍必须是 ErrUnauthorized（不泄露用户名是否注册）。
func TestLoginHidesUnknownUser(t *testing.T) {
	store := &authStore{findUserError: domain.ErrNotFound}
	useCases := service.NewAuthService(store, issuer{}, 30*24*time.Hour)

	_, err := useCases.Login(context.Background(), "ghost", "correct horse battery staple", "", "")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestRefreshRotatesToNewOpaqueToken(t *testing.T) {
	store := &authStore{user: repository.User{ID: 1, Role: "admin"}}
	useCases := service.NewAuthService(store, issuer{}, 30*24*time.Hour)
	result, err := useCases.Refresh(context.Background(), "old-token", "test", "ip-hash")
	require.NoError(t, err)
	require.NotEmpty(t, result.RefreshToken)
	require.Equal(t, security.TokenHash(result.RefreshToken), store.rotatedHash)
}
