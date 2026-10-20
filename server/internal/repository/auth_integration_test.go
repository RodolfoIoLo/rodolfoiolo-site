package repository_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	// Cleanup 是后进先出：先注册 Close，删数据的 Cleanup 才能在连接池还活着时执行。
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	store := repository.NewAuthRepository(pool)
	username := fmt.Sprintf("replay_%d", time.Now().UnixNano())
	user, err := store.CreateUser(ctx, username, "test-hash", "Replay Test", "", "admin")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
		require.NoError(t, err, "清理失败会留下脏数据，下一次运行必然撞唯一约束")
	})
	// token_hash 有全局唯一约束，写死常量会让第二次跑同一个库直接撞 23505。
	// 用 username 派生，保证每次运行互不冲突。
	hash := func(seed string) string {
		sum := sha256.Sum256([]byte(username + ":" + seed))
		return hex.EncodeToString(sum[:])
	}
	original, rotated, replayed, ipHash := hash("original"), hash("rotated"), hash("replayed"), hash("ip")

	require.NoError(t, store.CreateRefreshToken(ctx, user.ID, original, time.Now().Add(time.Hour), "test", ipHash))
	_, err = store.RotateRefreshToken(ctx, original, rotated, time.Now().Add(time.Hour), "test", ipHash)
	require.NoError(t, err)

	_, err = store.RotateRefreshToken(ctx, original, replayed, time.Now().Add(time.Hour), "test", ipHash)
	require.True(t, errors.Is(err, domain.ErrReplay))

	var active int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL`, user.ID).Scan(&active))
	require.Zero(t, active)
}
