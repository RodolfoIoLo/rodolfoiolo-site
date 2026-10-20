package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const projectSelect = `SELECT id,slug,name,summary,content,cover_image,technologies,repository_url,demo_url,featured,sort_order,status,started_at,completed_at,created_at,updated_at FROM projects`

type ProjectRepository struct {
	pool *pgxpool.Pool
}

// 返回指针:main.go里注入一次,所有handler共用同一个实例
func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		pool: pool,
	}
}

func (r *ProjectRepository) List(ctx context.Context, filter domain.ProjectFilter) (domain.Page[domain.Project], error) {
	conditions := make([]string, 0, 4) // 装SQL片段
	arguments := make([]any, 0, 4)     // 装对应的值
	// 4是容量预分配,预估最多4个条件,避免append时反复扩容,这只是性能提示
	add := func(format string, value any) {
		arguments = append(arguments, value)
		conditions = append(conditions, fmt.Sprintf(format, len(arguments)))
	}

	if filter.IncludePrivate {
		if filter.Status != "" {
			add("status=$%d", filter.Status) // 私有&按照指定状态筛选
		} else {
			conditions = append(conditions, "TRUE") // 私有&不筛选状态
		}
	} else {
		conditions = append(conditions, "status='active'") // 公开:硬编码只看active
	}

	if filter.Featured != nil {
		add("featured=$%d", filter.Featured)
	}

	if filter.Technology != "" {
		add("technologies @> to_jsonb(ARRAY[$%d]::text[])", filter.Technology)
	}
	// @> 是 PostgreSQL的包含运算符,用在JSONB上判断"左边数组包含右边数组"
	// to_jsonb(ARRAY[$3]::text[])把字符串转成JSON数组
	// 走GIN索引 projects_technologies_gin_idx（000002_init_content.up.sql:89）

	where := " WHERE " + strings.Join(conditions, " AND ") // strings.Join对空切片返回空字符串,"WHERE"之后没有内容PostgreSQL直接语法错误

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*)::int FROM projects"+where, arguments...).Scan(&total); err != nil { // 'arguments...`表示展开传进去
		return domain.Page[domain.Project]{}, fmt.Errorf("count projects: %w", err)
	}

	arguments = append(arguments, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.pool.Query(ctx, projectSelect+where+fmt.Sprintf(" ORDER BY featured DESC,sort_order,updated_at DESC LIMIT $%d OFFSET $%d", len(arguments)-1, len(arguments)), arguments...)
	if err != nil {
		return domain.Page[domain.Project]{}, fmt.Errorf("list projects: %w", err)
	}

	defer rows.Close()
	items := make([]domain.Project, 0, filter.PageSize)
	for rows.Next() {
		item, err := scanProject(rows)
		if err != nil {
			return domain.Page[domain.Project]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Page[domain.Project]{}, fmt.Errorf("iterate projects: %w", err)
	}
	return domain.Page[domain.Project]{Items: items, Page: filter.Page, PageSize: filter.PageSize, Total: total, TotalPages: pageCount(total, filter.PageSize)}, nil
}

func (r *ProjectRepository) Get(ctx context.Context, slug string, IncludePrivate bool) (domain.Project, error) {
	// IncludePrivate = false:公开接口,只能看到active
	// IncludePrivate = true:管理后台/内部调用,任何状态都能看
	condition := "slug=$1 AND status='active'"
	if IncludePrivate {
		condition = "slug=$1"
	}
	return scanProject(r.pool.QueryRow(ctx, projectSelect+" WHERE "+condition, slug))
}

func (r *ProjectRepository) Create(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	// technologies列式JSONB,Go里是[]string,库里是一段JSON文本
	technologies, err := json.Marshal(input.Technologies)
	if err != nil {
		return domain.Project{}, err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO projects(slug,name,summary,content,cover_image,technologies,repository_url,demo_url,featured,sort_order,status,started_at,completed_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, input.Slug, input.Name, input.Summary, input.Content, input.CoverImage, technologies, input.RepositoryURL, input.DemoURL, input.Featured, input.SortOrder, input.Status, input.StartedAt, input.CompletedAt)
	if isUniqueViolation(err) {
		return domain.Project{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("insert project: %w", err)
	}
	return r.Get(ctx, input.Slug, true)
}

func (r *ProjectRepository) Update(ctx context.Context, currentSlug string, input domain.ProjectInput) (domain.Project, error) {
	technologies, err := json.Marshal(input.Technologies)
	if err != nil {
		return domain.Project{}, err
	}
	command, err := r.pool.Exec(ctx, `UPDATE projects SET slug=$2,name=$3,summary=$4,content=$5,cover_image=$6,technologies=$7,
		repository_url=$8,demo_url=$9,featured=$10,sort_order=$11,status=$12,started_at=$13,completed_at=$14,updated_at=now() WHERE slug=$1`,
		currentSlug, input.Slug, input.Name, input.Summary, input.Content, input.CoverImage, technologies, input.RepositoryURL, input.DemoURL, input.Featured, input.SortOrder, input.Status, input.StartedAt, input.CompletedAt)
	if isUniqueViolation(err) {
		return domain.Project{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Project{}, domain.ErrNotFound
	}
	return r.Get(ctx, input.Slug, true)
}

func (r *ProjectRepository) Delete(ctx context.Context, slug string) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE slug=$1`, slug)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanProject(row rowScanner) (domain.Project, error) {
	var item domain.Project
	var technologies []byte

	err := row.Scan(&item.ID, &item.Slug, &item.Name, &item.Summary, &item.Content, &item.CoverImage, &technologies, &item.RepositoryURL, &item.DemoURL, &item.Featured, &item.SortOrder, &item.Status, &item.StartedAt, &item.CompletedAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("scan project: %w", err)
	}
	if err := json.Unmarshal(technologies, &item.Technologies); err != nil {
		return domain.Project{}, fmt.Errorf("decode technologies: %w", err)
	}
	return item, nil
}
