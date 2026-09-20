package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
	Email        *string
	AvatarURL    *string
	Bio          *string
	Role         string
	CreatedAt    time.Time
}

type AuthRepository struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool, now: time.Now}
}

func (r *AuthRepository) FindUserByUsername(ctx context.Context, username string) (User, error) {
	return scanUser(r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, nickname, email, avatar_url, bio, role, created_at
		FROM users WHERE username = $1`, username))
}

func (r *AuthRepository) FindUserByID(ctx context.Context, id int64) (User, error) {
	return scanUser(r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, nickname, email, avatar_url, bio, role, created_at
		FROM users WHERE id = $1`, id))
}

func (r *AuthRepository) CreateUser(ctx context.Context, username, passwordHash, nickname, email, role string) (User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, nickname, email, role)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id, username, password_hash, nickname, email, avatar_url, bio, role, created_at`,
		username, passwordHash, nickname, email, role))
	if isUniqueViolation(err) {
		return User{}, domain.ErrConflict
	}
	return user, err
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ipHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_hash)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))`,
		userID, tokenHash, expiresAt, userAgent, ipHash)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) RotateRefreshToken(ctx context.Context, currentHash, nextHash string, nextExpiresAt time.Time, userAgent, ipHash string) (User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return User{}, fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var tokenID, userID int64
	var expiresAt time.Time
	var revokedAt *time.Time
	var replacedBy *int64
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked_at, replaced_by
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE`, currentHash).Scan(&tokenID, &userID, &expiresAt, &revokedAt, &replacedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return User{}, fmt.Errorf("lock refresh token: %w", err)
	}

	if revokedAt != nil || replacedBy != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE refresh_tokens
			SET revoked_at = COALESCE(revoked_at, $2)
			WHERE user_id = $1`, userID, r.now().UTC()); err != nil {
			return User{}, fmt.Errorf("revoke replayed session family: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return User{}, fmt.Errorf("commit replay revocation: %w", err)
		}
		return User{}, domain.ErrReplay
	}

	if !expiresAt.After(r.now().UTC()) {
		if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1`, tokenID, r.now().UTC()); err != nil {
			return User{}, fmt.Errorf("revoke expired refresh token: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return User{}, fmt.Errorf("commit expired revocation: %w", err)
		}
		return User{}, domain.ErrUnauthorized
	}

	var nextID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_hash)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		RETURNING id`, userID, nextHash, nextExpiresAt, userAgent, ipHash).Scan(&nextID)
	if err != nil {
		return User{}, fmt.Errorf("insert rotated refresh token: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = $2, replaced_by = $3 WHERE id = $1`,
		tokenID, r.now().UTC(), nextID); err != nil {
		return User{}, fmt.Errorf("revoke previous refresh token: %w", err)
	}

	user, err := scanUser(tx.QueryRow(ctx, `
		SELECT id, username, password_hash, nickname, email, avatar_url, bio, role, created_at
		FROM users WHERE id = $1`, userID))
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit refresh rotation: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	command, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, $2)
		WHERE token_hash = $1`, tokenHash, r.now().UTC())
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}
	return nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanUser(row rowScanner) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.Email, &user.AvatarURL, &user.Bio, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, domain.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
