package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.slug, c.description, c.icon, c.parent_id, c.sort_order, c.created_at,
			count(a.id) FILTER (WHERE a.status = 'published')::int
		FROM categories c
		LEFT JOIN articles a ON a.category_id = c.id
		GROUP BY c.id
		ORDER BY c.sort_order, lower(c.name)`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	// 说明: var item []domain.Category同样可以声明变量,但是使用 ":=" 得到的是非nil的空切片,后者是nil,JSON序列化时nil切片变成NULL,空切片变成[],前端拿到null就得额外判空,所以API层统一用make(..., 0)
	items := make([]domain.Category, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (domain.Category, error) {
	return scanCategory(r.pool.QueryRow(ctx, `
		SELECT c.id, c.name, c.slug, c.description, c.icon, c.parent_id, c.sort_order, c.created_at,
		       count(a.id) FILTER (WHERE a.status = 'published')::int
		FROM categories c
		LEFT JOIN articles a ON a.category_id = c.id
		WHERE c.slug = $1 GROUP BY c.id`, slug))
}

func (r *CategoryRepository) Create(ctx context.Context, input domain.CategoryInput) (domain.Category, error) {
	item, err := scanCategory(r.pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug, description, icon, parent_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, slug, description, icon, parent_id, sort_order, created_at, 0`,
		input.Name, input.Slug, input.Description, input.Icon, input.ParentID, input.SortOrder))
	if isUniqueViolation(err) {
		return domain.Category{}, domain.ErrConflict
	}
	return item, err
}

func (r *CategoryRepository) Update(ctx context.Context, currentSlug string, input domain.CategoryInput) (domain.Category, error) {
	item, err := scanCategory(r.pool.QueryRow(ctx, `
		UPDATE categories SET name=$2, slug=$3, description=$4, icon=$5, parent_id=$6, sort_order=$7, updated_at=now()
		WHERE slug=$1
		RETURNING id, name, slug, description, icon, parent_id, sort_order, created_at,
		  (SELECT count(*)::int FROM articles WHERE category_id=categories.id AND status='published')`,
		currentSlug, input.Name, input.Slug, input.Description, input.Icon, input.ParentID, input.SortOrder))
	if isUniqueViolation(err) {
		return domain.Category{}, domain.ErrConflict
	}
	return item, err
}

func (r *CategoryRepository) Delete(ctx context.Context, slug string) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE slug=$1`, slug)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) ParentDepth(ctx context.Context, parentID int64) (int, error) {
	var parentParent *int64
	err := r.pool.QueryRow(ctx, `SELECT parent_id FROM categories WHERE id=$1`, parentID).Scan(&parentParent)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrBadRequest
	}
	if err != nil {
		return 0, fmt.Errorf("read category parent: %w", err)
	}
	if parentParent == nil {
		return 1, nil
	}
	return 2, nil
}

func scanCategory(row rowScanner) (domain.Category, error) {
	var item domain.Category
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Icon, &item.ParentID, &item.SortOrder, &item.CreatedAt, &item.ArticleCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Category{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("scan category: %w", err)
	}
	return item, nil
}
