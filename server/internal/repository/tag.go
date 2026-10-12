package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

func (r *TagRepository) List(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.name, t.slug, t.color, t.created_at,
			count(a.id) FILTER (WHERE a.status='published')::int
		FROM tags t
		LEFT JOIN article_tags at ON at.tag_id=t.id
		LEFT JOIN articles a ON a.id=at.article_id
		GROUP BY t.id ORDER BY lower(t.name)`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Tag, 0)
	for rows.Next() {
		item, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (domain.Tag, error) {
	return scanTag(r.pool.QueryRow(ctx, `
		SELECT t.id, t.name, t.slug, t.color, t.created_at,
		       count(a.id) FILTER (WHERE a.status='published')::int
		FROM tags t
		LEFT JOIN article_tags at ON at.tag_id=t.id
		LEFT JOIN articles a ON a.id=at.article_id
		WHERE t.slug=$1 GROUP BY t.id`, slug))
}

func (r *TagRepository) Create(ctx context.Context, input domain.TagInput) (domain.Tag, error) {
	item, err := scanTag(r.pool.QueryRow(ctx, `
		INSERT INTO tags (name, slug, color) VALUES ($1,$2,$3)
		RETURNING id,name,slug,color,created_at,0`, input.Name, input.Slug, input.Color))
	if isUniqueViolation(err) {
		return domain.Tag{}, domain.ErrConflict
	}
	return item, err
}

func (r *TagRepository) Update(ctx context.Context, currentSlug string, input domain.TagInput) (domain.Tag, error) {
	item, err := scanTag(r.pool.QueryRow(ctx, `
		UPDATE tags SET name=$2, slug=$3, color=$4, updated_at=now() WHERE slug=$1
		RETURNING id, name, slug, color, created_at,
		  (SELECT count(*)::int FROM article_tags at JOIN articles a ON a.id=at.article_id WHERE at.tag_id=tags.id AND a.status='published')`,
		currentSlug, input.Name, input.Slug, input.Color))
	if isUniqueViolation(err) {
		return domain.Tag{}, domain.ErrConflict
	}
	return item, err
}

func (r *TagRepository) Delete(ctx context.Context, slug string) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE slug=$1`, slug)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanTag(row rowScanner) (domain.Tag, error) {
	var item domain.Tag
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.Color, &item.CreatedAt, &item.ArticleCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Tag{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Tag{}, fmt.Errorf("scan tag: %w", err)
	}
	return item, nil
}
