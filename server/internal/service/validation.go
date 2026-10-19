package service

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	slugger "github.com/gosimple/slug"
)

var colorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func normalizedSlug(explicit, source string) (string, error) {
	value := strings.TrimSpace(explicit)
	if value == "" {
		value = slugger.Make(source)
	} else {
		value = slugger.Make(value)
	}
	if value == "" || len(value) > 200 {
		return "", domain.ErrBadRequest
	}
	return value, nil
}

func requiredText(value string, minimum, maximum int) (string, error) {
	value = strings.TrimSpace(value)
	length := utf8.RuneCountInString(value)
	if length < minimum || length > maximum {
		return "", domain.ErrBadRequest
	}
	return value, nil
}

func optionalHTTPURL(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, domain.ErrBadRequest
	}
	return &trimmed, nil
}

func normalizedTechnologies(values []string) ([]string, error) {
	if len(values) > 20 {
		return nil, domain.ErrBadRequest
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || utf8.RuneCountInString(value) > 50 {
			return nil, domain.ErrBadRequest
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func normalizedPage(page, pageSize *int) (int, int, error) {
	resultPage, resultSize := 1, 20
	if page != nil && *page != 0 {
		resultPage = *page
	}
	if pageSize != nil && *pageSize != 0 {
		resultSize = *pageSize
	}
	if resultPage < 1 || resultSize < 1 || resultSize > 100 {
		return 0, 0, domain.ErrBadRequest
	}
	return resultPage, resultSize, nil
}

func validColor(color *string) error {
	if color != nil && !colorPattern.MatchString(*color) {
		return domain.ErrBadRequest
	}
	return nil
}

func statusAllowed(value string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return errors.Join(domain.ErrBadRequest, errors.New("invalid status"))
}
