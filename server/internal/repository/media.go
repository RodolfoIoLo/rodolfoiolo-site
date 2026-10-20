package repository

import (
	"context"
	"fmt"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct{ pool *pgxpool.Pool }

func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository { return &MediaRepository{pool: pool} }
func (r *MediaRepository) Create(ctx context.Context, item domain.MediaAsset, userID int64) (domain.MediaAsset, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO media_assets(object_key,original_name,mime_type,byte_size,width,height,sha256,uploaded_by)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,created_at`, item.ObjectKey, item.OriginalName, item.MIMEType, item.ByteSize, item.Width, item.Height, item.SHA256, userID).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return domain.MediaAsset{}, fmt.Errorf("insert media metadata: %w", err)
	}
	return item, nil
}
