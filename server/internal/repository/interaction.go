package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InteractionRepository struct{ pool *pgxpool.Pool }

func NewInteractionRepository(pool *pgxpool.Pool) *InteractionRepository {
	return &InteractionRepository{pool: pool}
}

func (r *InteractionRepository) ListComments(ctx context.Context, articleSlug string, page, pageSize int, includePending bool, status string) (domain.Page[domain.Comment], error) {
	conditions := "a.slug=$1 AND c.status='approved'"
	arguments := []any{articleSlug}

	if includePending {
		conditions = "a.slug=$1"
	}

	if status != "" {
		arguments = append(arguments, status)
		conditions += " AND c.status=$2"
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)::int FROM comments c JOIN articles a ON a.id=c.article_id WHERE `+conditions, arguments...).Scan(&total); err != nil {
		return domain.Page[domain.Comment]{}, fmt.Errorf("count comments: %w", err)
	}

	arguments = append(arguments, pageSize, (page-1)*pageSize)
	rows, err := r.pool.Query(ctx, `SELECT c.id,c.article_id,c.parent_id,c.author_name,c.content,c.status,c.created_at FROM comments c JOIN articles a ON a.id=c.article_id WHERE `+conditions+fmt.Sprintf(" ORDER BY c.created_at LIMIT $%d OFFSET $%d", len(arguments)-1, len(arguments)), arguments...)
	if err != nil {
		return domain.Page[domain.Comment]{}, fmt.Errorf("list comments: %w", err)
	}

	defer rows.Close()

	items := make([]domain.Comment, 0, pageSize)
	for rows.Next() {
		var item domain.Comment
		if err := rows.Scan(&item.ID, &item.ArticleID, &item.ParentID, &item.Nickname, &item.Content, &item.Status, &item.CreatedAt); err != nil {
			return domain.Page[domain.Comment]{}, err
		}
		items = append(items, item)
	}
	return domain.Page[domain.Comment]{Items: items, Page: page, PageSize: pageSize, Total: total, TotalPages: pageCount(total, pageSize)}, rows.Err()
}

func (r *InteractionRepository) ListAdminComments(ctx context.Context, page, pageSize int, status string) (domain.Page[domain.Comment], error) {
	condition := "TRUE"
	arguments := []any{}

	if status != "" {
		condition = "status=$1"
		arguments = append(arguments, status)
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*)::int FROM comments WHERE "+condition, arguments...).Scan(&total); err != nil {
		return domain.Page[domain.Comment]{}, err
	}

	arguments = append(arguments, pageSize, (page-1)*pageSize)
	rows, err := r.pool.Query(ctx, `SELECT id,article_id,parent_id,author_name,content,status,created_at FROM comments WHERE `+condition+fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(arguments)-1, len(arguments)), arguments...)
	if err != nil {
		return domain.Page[domain.Comment]{}, err
	}
	defer rows.Close()

	items := make([]domain.Comment, 0, pageSize)
	for rows.Next() {
		var item domain.Comment
		if err := rows.Scan(&item.ID, &item.ArticleID, &item.ParentID, &item.Nickname, &item.Content, &item.Status, &item.CreatedAt); err != nil {
			return domain.Page[domain.Comment]{}, err
		}
		items = append(items, item)
	}
	return domain.Page[domain.Comment]{Items: items, Page: page, PageSize: pageSize, Total: total, TotalPages: pageCount(total, pageSize)}, rows.Err()
}

func (r *InteractionRepository) CreateComment(ctx context.Context, input domain.CommentInput) (domain.Comment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Comment{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	var articleID int64
	if err := tx.QueryRow(ctx, `SELECT id FROM articles WHERE slug=$1 AND status='published' FOR SHARE`, input.ArticleSlug).Scan(&articleID); errors.Is(err, pgx.ErrNoRows) {
		return domain.Comment{}, domain.ErrNotFound
	} else if err != nil {
		return domain.Comment{}, err
	}
	if input.ParentID != nil {
		var parentArticle int64
		var grandParent *int64
		err := tx.QueryRow(ctx, `SELECT article_id,parent_id FROM comments WHERE id=$1 AND status='approved' FOR SHARE`, *input.ParentID).Scan(&parentArticle, &grandParent)
		if errors.Is(err, pgx.ErrNoRows) || parentArticle != articleID || grandParent != nil {
			return domain.Comment{}, domain.ErrBadRequest
		}
		if err != nil {
			return domain.Comment{}, err
		}
	}
	var item domain.Comment
	err = tx.QueryRow(ctx, `INSERT INTO comments(article_id,parent_id,author_name,author_email_hash,content,visitor_id,ip_hash,user_agent)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,article_id,parent_id,author_name,content,status,created_at`, articleID, input.ParentID, input.Nickname, input.EmailHash, input.Content, input.VisitorID, input.IPHash, input.UserAgent).Scan(&item.ID, &item.ArticleID, &item.ParentID, &item.Nickname, &item.Content, &item.Status, &item.CreatedAt)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("insert comment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, err
	}
	return item, nil
}

func (r *InteractionRepository) ReviewComment(ctx context.Context, id, userID int64, status string) (domain.Comment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Comment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var item domain.Comment
	if err := tx.QueryRow(ctx, `SELECT id,article_id,parent_id,author_name,content,status,created_at FROM comments WHERE id=$1 FOR UPDATE`, id).Scan(&item.ID, &item.ArticleID, &item.ParentID, &item.Nickname, &item.Content, &item.Status, &item.CreatedAt); errors.Is(err, pgx.ErrNoRows) {
		return domain.Comment{}, domain.ErrNotFound
	} else if err != nil {
		return domain.Comment{}, err
	}
	previous := item.Status
	if _, err := tx.Exec(ctx, `UPDATE comments SET status=$2,reviewed_by=$3,reviewed_at=now() WHERE id=$1`, id, status, userID); err != nil {
		return domain.Comment{}, err
	}
	item.Status = status
	if previous != "approved" && status == "approved" {
		_, err = tx.Exec(ctx, `UPDATE articles SET comment_count=comment_count+1 WHERE id=$1`, item.ArticleID)
	} else if previous == "approved" && status != "approved" {
		_, err = tx.Exec(ctx, `UPDATE articles SET comment_count=GREATEST(comment_count-1,0) WHERE id=$1`, item.ArticleID)
	}
	if err != nil {
		return domain.Comment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, err
	}
	return item, nil
}

func (r *InteractionRepository) DeleteComment(ctx context.Context, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var articleID int64
	var status string
	err = tx.QueryRow(ctx, `DELETE FROM comments WHERE id=$1 RETURNING article_id,status`, id).Scan(&articleID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if status == "approved" {
		if _, err = tx.Exec(ctx, `UPDATE articles SET comment_count=GREATEST(comment_count-1,0) WHERE id=$1`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *InteractionRepository) SetLike(ctx context.Context, articleSlug, visitorID, ipHash string, liked bool) (bool, int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var articleID int64
	if err := tx.QueryRow(ctx, `SELECT id FROM articles WHERE slug=$1 AND status='published' FOR UPDATE`, articleSlug).Scan(&articleID); errors.Is(err, pgx.ErrNoRows) {
		return false, 0, domain.ErrNotFound
	} else if err != nil {
		return false, 0, err
	}
	if liked {
		command, err := tx.Exec(ctx, `INSERT INTO likes(article_id,visitor_id,ip_hash) VALUES($1,$2,$3) ON CONFLICT(article_id,visitor_id) DO NOTHING`, articleID, visitorID, ipHash)
		if err != nil {
			return false, 0, err
		}
		if command.RowsAffected() > 0 {
			_, err = tx.Exec(ctx, `UPDATE articles SET like_count=like_count+1 WHERE id=$1`, articleID)
			if err != nil {
				return false, 0, err
			}
		}
	} else {
		command, err := tx.Exec(ctx, `DELETE FROM likes WHERE article_id=$1 AND visitor_id=$2`, articleID, visitorID)
		if err != nil {
			return false, 0, err
		}
		if command.RowsAffected() > 0 {
			_, err = tx.Exec(ctx, `UPDATE articles SET like_count=GREATEST(like_count-1,0) WHERE id=$1`, articleID)
			if err != nil {
				return false, 0, err
			}
		}
	}
	var finalLiked bool
	var count int
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE article_id=$1 AND visitor_id=$2),like_count::int FROM articles WHERE id=$1`, articleID, visitorID).Scan(&finalLiked, &count); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, 0, err
	}
	return finalLiked, count, nil
}

func (r *InteractionRepository) LikeStatus(ctx context.Context, articleSlug, visitorID string) (bool, int, error) {
	var liked bool
	var count int
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes l WHERE l.article_id=a.id AND l.visitor_id=$2),a.like_count::int FROM articles a WHERE a.slug=$1 AND a.status='published'`, articleSlug, visitorID).Scan(&liked, &count)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, 0, domain.ErrNotFound
	}
	return liked, count, err
}
