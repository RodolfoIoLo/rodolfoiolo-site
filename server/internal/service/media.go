package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/storage"
)

const maximumImageBytes = 6 << 20

type MediaResult struct {
	Asset domain.MediaAsset
	URL   string
}
type MediaService struct {
	repository *repository.MediaRepository
	store      storage.ObjectStore
}

func NewMediaService(repository *repository.MediaRepository, store storage.ObjectStore) *MediaService {
	return &MediaService{repository: repository, store: store}
}
func (s *MediaService) Upload(ctx context.Context, userID int64, filename string, data []byte) (MediaResult, error) {
	if len(data) == 0 || len(data) > maximumImageBytes {
		return MediaResult{}, domain.ErrBadRequest
	}
	mime := http.DetectContentType(data)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	extension, ok := extensions[mime]
	if !ok {
		return MediaResult{}, domain.ErrBadRequest
	}
	configuration, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || configuration.Width < 1 || configuration.Height < 1 || configuration.Width > 8000 || configuration.Height > 8000 {
		return MediaResult{}, domain.ErrBadRequest
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return MediaResult{}, err
	}
	key := "media/" + hex.EncodeToString(random) + extension
	sum := sha256.Sum256(data)
	if err := s.store.Put(ctx, key, mime, data); err != nil {
		return MediaResult{}, err
	}
	asset, err := s.repository.Create(ctx, domain.MediaAsset{ObjectKey: key, OriginalName: filepath.Base(strings.ReplaceAll(filename, "\\", "/")), MIMEType: mime, ByteSize: int64(len(data)), Width: configuration.Width, Height: configuration.Height, SHA256: hex.EncodeToString(sum[:])}, userID)
	if err != nil {
		deleteErr := s.store.Delete(ctx, key)
		if deleteErr != nil {
			return MediaResult{}, errors.Join(err, deleteErr)
		}
		return MediaResult{}, err
	}
	return MediaResult{Asset: asset, URL: s.store.PublicURL(key)}, nil
}
