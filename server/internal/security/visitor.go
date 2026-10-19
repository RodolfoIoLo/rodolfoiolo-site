package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// VisitorSigner 的作用:
// 1. 生成一个"访客标识", 用HMAC-SHA256对这个标识做签名, 把"id + 签名"拼成一个可验证的token, 之后再校验这个token是否被篡改过
type VisitorSigner struct {
	secret []byte
}

func NewVisitorSigner(secret []byte) (VisitorSigner, error) {
	if len(secret) < 32 {
		return VisitorSigner{}, errors.New("visitor secret must contain at least 32 bytes")
	}
	return VisitorSigner{secret: append([]byte(nil), secret...)}, nil
}

// 生成新访客token
func (s VisitorSigner) New() (string, string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", err
	}
	id := base64.RawURLEncoding.EncodeToString(buffer)
	return id, id + "." + s.signature(id), nil
}

// 验证token
func (s VisitorSigner) Verify(value string) (string, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" {
		return "", false
	}
	expected := s.signature(parts[0])
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return "", false
	}
	return parts[0], true
}

// 私有签名函数
func (s VisitorSigner) signature(id string) string {
	digest := hmac.New(sha256.New, s.secret)
	_, _ = digest.Write([]byte(id))
	return base64.RawURLEncoding.EncodeToString(digest.Sum(nil))
}
