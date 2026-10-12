package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SiteRepository struct {
	pool *pgxpool.Pool
}

func NewSiteRepository(pool *pgxpool.Pool) *SiteRepository {
	return &SiteRepository{pool: pool}
}

func (r *SiteRepository) GetSettinngs(ctx context.Context) (domain.SiteSettings, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT value FROM site_settings WHERE key='site' AND is_public=true`).Scan(&raw)
	if err != nil {
		return domain.SiteSettings{}, fmt.Errorf("read site settings: %w", err)
	}
	var settings domain.SiteSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return domain.SiteSettings{}, fmt.Errorf("decode site settings: %w", err)
	}
	return settings, nil
}

func (r *SiteRepository) UPdateSettings(ctx context.Context, settings domain.SiteSettings, userID int64) error {
	raw, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO site_settings(key,value,is_public,updated_by,updated_at) VALUES('site',$1,true,$2,now())
		ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value,is_public=true,updated_by=EXCLUDED.updated_by,updated_at=now()`, raw, userID)
	if err != nil {
		return fmt.Errorf("update site settings: %w", err)
	}
	return nil
}

func (r *SiteRepository) PublicStats(ctx context.Context) (articles, projects, categories, tags int, err error) {
	err = r.pool.QueryRow(ctx, `SELECT
		(SELECT count(*)::int FROM articles WHERE status='published'),
		(SELECT count(*)::int FROM projects WHERE status='active'),
		(SELECT count(*)::int FROM categories),
		(SELECT count(*)::int FROM tags)`).Scan(&articles, &projects, &categories, &tags)
	return
}

func (r *SiteRepository) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	var result domain.Dashboard
	err := r.pool.QueryRow(ctx, `SELECT
		(SELECT count(*)::int FROM articles),
		(SELECT count(*)::int FROM articles WHERE status='draft'),
		(SELECT count(*)::int FROM projects),
		(SELECT count(*)::int FROM comments WHERE status='pending'),
		(SELECT COALESCE(sum(view_count),0)::int FROM articles),
		(SELECT COALESCE(sum(like_count),0)::int FROM articles)`).Scan(&result.TotalArticles, &result.DraftArticles, &result.TotalProjects, &result.PendingComments, &result.TotalViews, &result.TotalLikes)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("read dashboard: %w", err)
	}
	return result, nil
}
