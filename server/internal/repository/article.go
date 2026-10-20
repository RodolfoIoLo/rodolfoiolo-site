package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const articleSelect = `
	SELECT a.id,a.slug,a.title,a.summary,a.content,a.cover_image,a.status,a.is_top,
	       a.view_count::int,a.like_count::int,a.comment_count::int,a.published_at,a.created_at,a.updated_at,
	       u.id,u.username,u.nickname,u.avatar_url,u.bio,u.role,u.created_at,
	       c.id,c.name,c.slug,c.description,c.icon,c.parent_id,c.sort_order,c.created_at,
		   COALESCE((
		     SELECT jsonb_agg(jsonb_build_object(
	           'id',t.id,'name',t.name,'slug',t.slug,'color',t.color,'created_at',t.created_at
	         ) ORDER BY lower(t.name))
	         FROM article_tags at JOIN tags t ON t.id=at.tag_id WHERE at.article_id=a.id
	       ), '[]'::jsonb)
	FROM articles a
	JOIN users u ON u.id=a.author_id
	LEFT JOIN categories c ON c.id=a.category_id`

type ArticleRepository struct {
	pool *pgxpool.Pool
}

func NewArticleRepository(pool *pgxpool.Pool) *ArticleRepository {
	return &ArticleRepository{pool: pool}
}

func (r *ArticleRepository) List(ctx context.Context, filter domain.ArticleFilter) (domain.Page[domain.Article], error) {
	// 把筛选条件翻译成 WHERE 片段 + 参数数组
	conditions, arguments := articleConditions(filter)
	// 用 AND 拼接起来
	where := " WHERE " + strings.Join(conditions, " AND ")

	// 先算总页数(分页要)
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*)::int FROM articles a"+where, arguments...).Scan(&total); err != nil {
		return domain.Page[domain.Article]{}, fmt.Errorf("count articles: %w", err)
	}
	order := "a.is_top DESC, a.published_at DESC NULLS LAST, a.created_at DESC"
	// 排序只认白名单,不接收用户原文
	if filter.Sort == "hottest" {
		order = "a.is_top DESC, a.like_count DESC, a.published_at DESC NULLS LAST"
	}
	if filter.Sort == "views" {
		order = "a.is_top DESC, a.view_count DESC, a.published_at DESC NULLS LAST"
	}
	// 追加 LIMIT / OFFSET 参数
	arguments = append(arguments, filter.PageSize, (filter.Page-1)*filter.PageSize)
	query := articleSelect + where + fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", order, len(arguments)-1, len(arguments))
	// 真正取数据
	rows, err := r.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.Page[domain.Article]{}, fmt.Errorf("list articles: %w", err)
	}
	// 必须关,否则连接不还给池
	defer rows.Close()
	items := make([]domain.Article, 0, filter.PageSize)
	for rows.Next() {
		// 每行扫一次
		item, err := scanArticle(rows)
		if err != nil {
			return domain.Page[domain.Article]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Page[domain.Article]{}, fmt.Errorf("iterate articles: %w", err)
	}
	return domain.Page[domain.Article]{Items: items, Page: filter.Page, PageSize: filter.PageSize, Total: total, TotalPages: pageCount(total, filter.PageSize)}, nil
}

func (r *ArticleRepository) Get(ctx context.Context, slug string, IncludePrivate bool) (domain.Article, error) {
	condition := "a.slug=$1 AND a.status='published'"
	if IncludePrivate {
		condition = "a.slug=$1"
	}
	return scanArticle(r.pool.QueryRow(ctx, articleSelect+" WHERE "+condition, slug))
}

func (r *ArticleRepository) Create(ctx context.Context, authorID int64, input domain.ArticleInput) (domain.Article, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Article{}, fmt.Errorf("begin article creation: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	var publishedAt *time.Time
	if input.Status == "published" {
		now := time.Now().UTC()
		publishedAt = &now
	}
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO articles (slug,title,summary,content,cover_image,category_id,author_id,status,is_top,published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		input.Slug, input.Title, input.Summary, input.Content, input.CoverImage, input.CategoryID, authorID, input.Status, input.IsTop, publishedAt).Scan(&id)
	if isUniqueViolation(err) {
		return domain.Article{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("insert article: %w", err)
	}
	if err := replaceArticleTags(ctx, tx, id, input.TagIDs); err != nil {
		return domain.Article{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Article{}, fmt.Errorf("commit article creation: %w", err)
	}
	return r.Get(ctx, input.Slug, true)
}

func (r *ArticleRepository) Update(ctx context.Context, currentSlug string, input domain.ArticleInput) (domain.Article, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Article{}, fmt.Errorf("begin article update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id int64
	err = tx.QueryRow(ctx, `
		UPDATE articles SET slug=$2,title=$3,summary=$4,content=$5,cover_image=$6,category_id=$7,is_top=$8,updated_at=now()
		WHERE slug=$1 RETURNING id`, currentSlug, input.Slug, input.Title, input.Summary, input.Content, input.CoverImage, input.CategoryID, input.IsTop).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Article{}, domain.ErrNotFound
	}
	if isUniqueViolation(err) {
		return domain.Article{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("update article: %w", err)
	}
	if err := replaceArticleTags(ctx, tx, id, input.TagIDs); err != nil {
		return domain.Article{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Article{}, fmt.Errorf("commit article update: %w", err)
	}
	return r.Get(ctx, input.Slug, true)
}

func (r *ArticleRepository) SetStatus(ctx context.Context, slug, status string) (domain.Article, error) {
	command, err := r.pool.Exec(ctx, `
		UPDATE articles SET status=$2,
		  published_at=CASE WHEN $2='published' THEN COALESCE(published_at,now()) ELSE NULL END,
		  updated_at=now() WHERE slug=$1`, slug, status)
	if err != nil {
		return domain.Article{}, fmt.Errorf("set article status: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Article{}, domain.ErrNotFound
	}
	return r.Get(ctx, slug, true)
}

func (r *ArticleRepository) SetTop(ctx context.Context, slug string, top bool) (domain.Article, error) {
	command, err := r.pool.Exec(ctx, `UPDATE articles SET is_top=$2,updated_at=now() WHERE slug=$1`, slug, top)
	if err != nil {
		return domain.Article{}, fmt.Errorf("set article top: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Article{}, domain.ErrNotFound
	}
	return r.Get(ctx, slug, true)
}

func (r *ArticleRepository) Delete(ctx context.Context, slug string) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM articles WHERE slug=$1`, slug)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ArticleRepository) IncrementView(ctx context.Context, slug string) error {
	_, err := r.pool.Exec(ctx, `UPDATE articles SET view_count=view_count+1 WHERE slug=$1 AND status='published'`, slug)
	return err
}

func articleConditions(filter domain.ArticleFilter) ([]string, []any) {
	conditions := make([]string, 0, 5)
	arguments := make([]any, 0, 5)
	add := func(format string, value any) {
		arguments = append(arguments, value)
		conditions = append(conditions, fmt.Sprintf(format, len(arguments)))
	}
	if filter.IncludePrivate {
		if filter.Status != "" {
			add("a.status=$%d", filter.Status)
		} else {
			conditions = append(conditions, "TRUE")
		}
	} else {
		conditions = append(conditions, "a.status='published'")
	}
	if filter.Category != "" {
		add("EXISTS (SELECT 1 FROM categories c WHERE c.id=a.category_id AND c.slug=$%d)", filter.Category)
	}
	if filter.Tag != "" {
		add("EXISTS (SELECT 1 FROM article_tags at JOIN tags t ON t.id=at.tag_id WHERE at.article_id=a.id AND t.slug=$%d)", filter.Tag)
	}
	if filter.Keyword != "" {
		arguments = append(arguments, "%"+escapeLike(filter.Keyword)+"%")
		conditions = append(conditions, fmt.Sprintf("(a.title ILIKE $%d ESCAPE E'\\\\' OR a.summary ILIKE $%d ESCAPE E'\\\\')", len(arguments), len(arguments)))
	}
	return conditions, arguments
}

func replaceArticleTags(ctx context.Context, tx pgx.Tx, articleID int64, tagIDs []int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM article_tags WHERE article_id=$1`, articleID); err != nil {
		return fmt.Errorf("clear article tags: %w", err)
	}
	if len(tagIDs) == 0 {
		return nil
	}
	command, err := tx.Exec(ctx, `
		INSERT INTO article_tags(article_id,tag_id)
		SELECT $1,id FROM tags WHERE id=ANY($2::bigint[])`, articleID, tagIDs)
	if err != nil {
		return fmt.Errorf("insert article tags: %w", err)
	}
	if command.RowsAffected() != int64(len(tagIDs)) {
		return domain.ErrBadRequest
	}
	return nil
}

func scanArticle(row rowScanner) (domain.Article, error) {
	var item domain.Article
	var categoryID *int64
	var categoryName, categorySlug *string
	var categoryDescription, categoryIcon *string
	var categoryParent *int64
	var categorySort *int
	var categoryCreated *time.Time
	var tagsJSON []byte
	err := row.Scan(&item.ID, &item.Slug, &item.Title, &item.Summary, &item.Content, &item.CoverImage, &item.Status, &item.IsTop,
		&item.ViewCount, &item.LikeCount, &item.CommentCount, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt,
		&item.Author.ID, &item.Author.Username, &item.Author.Nickname, &item.Author.AvatarURL, &item.Author.Bio, &item.Author.Role, &item.Author.CreatedAt,
		&categoryID, &categoryName, &categorySlug, &categoryDescription, &categoryIcon, &categoryParent, &categorySort, &categoryCreated, &tagsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Article{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("scan article: %w", err)
	}
	if categoryID != nil {
		item.Category = &domain.Category{ID: *categoryID, Name: *categoryName, Slug: *categorySlug, Description: categoryDescription, Icon: categoryIcon, ParentID: categoryParent, SortOrder: *categorySort, CreatedAt: *categoryCreated}
	}
	var tags []struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		Slug      string    `json:"slug"`
		Color     *string   `json:"color"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		return domain.Article{}, fmt.Errorf("decode article tags: %w", err)
	}
	item.Tags = make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		item.Tags = append(item.Tags, domain.Tag{ID: tag.ID, Name: tag.Name, Slug: tag.Slug, Color: tag.Color, CreatedAt: tag.CreatedAt})
	}
	return item, nil
}

func pageCount(total, size int) int {
	if total == 0 {
		return 0
	}
	return (total + size - 1) / size
}
func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}
