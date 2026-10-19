package security

import (
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

type CommentSanitizer struct {
	policy *bluemonday.Policy
}

func NewCommentSanitizer() *CommentSanitizer {
	policy := bluemonday.StrictPolicy()
	return &CommentSanitizer{policy: policy}
}

func (s *CommentSanitizer) Clean(value string) string {
	plain := s.policy.Sanitize(value)
	plain = html.UnescapeString(plain)
	return strings.TrimSpace(plain)
}
