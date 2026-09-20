package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestRefreshReplayRevokesSessionFamily(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	store := repository.NewAuthRepository(pool)
	username := fmt.Sprintf("replay_%d", time.Now().UnixNano())
	user, err := store.CreateUser(ctx, username, "test-hash", "Replay Test", "", "admin")
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	require.NoError(t, store.CreateRefreshToken(ctx, user.ID, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", time.Now().Add(time.Hour), "test", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
	_, err = store.RotateRefreshToken(ctx,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		time.Now().Add(time.Hour), "test", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NoError(t, err)

	_, err = store.RotateRefreshToken(ctx,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		time.Now().Add(time.Hour), "test", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.True(t, errors.Is(err, domain.ErrReplay))

	var active int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL`, user.ID).Scan(&active))
	require.Zero(t, active)
}
