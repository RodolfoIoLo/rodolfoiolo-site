package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
)

type InteractionService struct {
	repository *repository.InteractionRepository
	sanitizer  *security.CommentSanitizer
	secret     []byte
}

func NewInteractionService(repository *repository.InteractionRepository, sanitizer *security.CommentSanitizer, secret []byte) *InteractionService {
	return &InteractionService{repository: repository, sanitizer: sanitizer, secret: append([]byte(nil), secret...)}
}
func (s *InteractionService) List(ctx context.Context, slug string, page, pageSize int) (domain.Page[domain.Comment], error) {
	normalizedPage, normalizedSize, err := normalizedPage(&page, &pageSize)
	if err != nil {
		return domain.Page[domain.Comment]{}, err
	}
	return s.repository.ListComments(ctx, slug, normalizedPage, normalizedSize, false, "")
}
func (s *InteractionService) ListAdmin(ctx context.Context, page, pageSize int, status string) (domain.Page[domain.Comment], error) {
	normalizedPage, normalizedSize, err := normalizedPage(&page, &pageSize)
	if err != nil {
		return domain.Page[domain.Comment]{}, err
	}
	if status != "" {
		if err := statusAllowed(status, "pending", "approved", "rejected"); err != nil {
			return domain.Page[domain.Comment]{}, err
		}
	}
	return s.repository.ListAdminComments(ctx, normalizedPage, normalizedSize, status)
}
func (s *InteractionService) Create(ctx context.Context, input domain.CommentInput, email, honeypot string) (domain.Comment, error) {
	if strings.TrimSpace(honeypot) != "" {
		return domain.Comment{}, domain.ErrBadRequest
	}
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Content = s.sanitizer.Clean(input.Content)
	if utf8.RuneCountInString(input.Nickname) < 1 || utf8.RuneCountInString(input.Nickname) > 100 || utf8.RuneCountInString(input.Content) < 2 || utf8.RuneCountInString(input.Content) > 5000 || input.VisitorID == "" {
		return domain.Comment{}, domain.ErrBadRequest
	}
	if email != "" {
		hash := security.HMACHash(s.secret, strings.ToLower(strings.TrimSpace(email)))
		input.EmailHash = &hash
	}
	return s.repository.CreateComment(ctx, input)
}
func (s *InteractionService) Review(ctx context.Context, id, userID int64, status string) (domain.Comment, error) {
	if err := statusAllowed(status, "approved", "rejected"); err != nil {
		return domain.Comment{}, err
	}
	return s.repository.ReviewComment(ctx, id, userID, status)
}
func (s *InteractionService) Delete(ctx context.Context, id int64) error {
	return s.repository.DeleteComment(ctx, id)
}
func (s *InteractionService) SetLike(ctx context.Context, slug, visitorID, ipHash string, liked bool) (bool, int, error) {
	if visitorID == "" {
		return false, 0, domain.ErrBadRequest
	}
	return s.repository.SetLike(ctx, slug, visitorID, ipHash, liked)
}
func (s *InteractionService) LikeStatus(ctx context.Context, slug, visitorID string) (bool, int, error) {
	if visitorID == "" {
		return false, 0, domain.ErrBadRequest
	}
	return s.repository.LikeStatus(ctx, slug, visitorID)
}
