package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

type AuthStore interface {
	FindUserByUsername(context.Context, string) (repository.User, error)
	FindUserByID(context.Context, int64) (repository.User, error)
	CreateRefreshToken(context.Context, int64, string, time.Time, string, string) error
	RotateRefreshToken(context.Context, string, string, time.Time, string, string) (repository.User, error)
	RevokeRefreshToken(context.Context, string) error
}

type AccessIssuer interface {
	Issue(security.Identity) (string, time.Time, error)
	TTL() time.Duration
}

type AuthResult struct {
	AccessToken  string
	ExpiresIn    int
	RefreshToken string
	User         repository.User
}

type AuthService struct {
	store      AuthStore
	issuer     AccessIssuer
	refreshTTL time.Duration
	now        func() time.Time
}

func NewAuthService(store AuthStore, issuer AccessIssuer, refreshTTL time.Duration) *AuthService {
	return &AuthService{store: store, issuer: issuer, refreshTTL: refreshTTL, now: time.Now}
}

func (s *AuthService) Login(ctx context.Context, username, password, userAgent, ipHash string) (AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return AuthResult{}, domain.ErrBadRequest
	}
	user, err := s.store.FindUserByUsername(ctx, username)
	if err != nil {
		// 用户不存在折叠成 ErrUnauthorized，不把「这个用户名没注册」暴露给调用方。
		if errors.Is(err, domain.ErrNotFound) {
			return AuthResult{}, domain.ErrUnauthorized
		}
		// 其余错误（连接失败、超时、SQL 错误）必须原样上抛。
		//
		// 这里的顺序不能颠倒：若把 err 判断排在 VerifyPassword 之后，
		// 出错时 user 是零值、PasswordHash 是空串，bcrypt 解析空哈希失败返回 false，
		// 于是数据库故障会被当成「密码错误」返回 401，日志里不留任何痕迹。
		return AuthResult{}, err
	}
	if !security.VerifyPassword(user.PasswordHash, password) {
		return AuthResult{}, domain.ErrUnauthorized
	}
	return s.issueSession(ctx, user, userAgent, ipHash)
}

func (s *AuthService) Refresh(ctx context.Context, currentRaw, userAgent, ipHash string) (AuthResult, error) {
	if currentRaw == "" {
		return AuthResult{}, domain.ErrUnauthorized
	}
	nextRaw, err := security.NewRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}
	user, err := s.store.RotateRefreshToken(ctx, security.TokenHash(currentRaw), security.TokenHash(nextRaw), s.now().UTC().Add(s.refreshTTL), userAgent, ipHash)
	if err != nil {
		return AuthResult{}, err
	}
	access, _, err := s.issuer.Issue(security.Identity{UserID: user.ID, Role: user.Role})
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{AccessToken: access, ExpiresIn: int(s.issuer.TTL().Seconds()), RefreshToken: nextRaw, User: user}, nil
}

func (s *AuthService) Logout(ctx context.Context, currentRaw string) error {
	if currentRaw == "" {
		return domain.ErrUnauthorized
	}
	return s.store.RevokeRefreshToken(ctx, security.TokenHash(currentRaw))
}

func (s *AuthService) CurrentUser(ctx context.Context, userID int64) (repository.User, error) {
	if userID <= 0 {
		return repository.User{}, domain.ErrUnauthorized
	}
	return s.store.FindUserByID(ctx, userID)
}

func (s *AuthService) issueSession(ctx context.Context, user repository.User, userAgent, ipHash string) (AuthResult, error) {
	refreshRaw, err := security.NewRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.store.CreateRefreshToken(ctx, user.ID, security.TokenHash(refreshRaw), s.now().UTC().Add(s.refreshTTL), userAgent, ipHash); err != nil {
		return AuthResult{}, err
	}
	access, _, err := s.issuer.Issue(security.Identity{UserID: user.ID, Role: user.Role})
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{AccessToken: access, ExpiresIn: int(s.issuer.TTL().Seconds()), RefreshToken: refreshRaw, User: user}, nil
}
