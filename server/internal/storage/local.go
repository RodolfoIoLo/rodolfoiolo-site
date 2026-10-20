package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Local struct{ root, publicBase string }

func NewLocal(root, publicBase string) (*Local, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o750); err != nil {
		return nil, err
	}
	return &Local{root: absolute, publicBase: strings.TrimRight(publicBase, "/")}, nil
}
func (s *Local) Put(_ context.Context, key, _ string, data []byte) error {
	path, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("write local object: %w", err)
	}
	return nil
}
func (s *Local) Delete(_ context.Context, key string) error {
	path, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func (s *Local) PublicURL(key string) string { return s.publicBase + "/" + strings.TrimLeft(key, "/") }
func (s *Local) safePath(key string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", errors.New("invalid object key")
	}
	path := filepath.Join(s.root, clean)
	if !strings.HasPrefix(path, s.root+string(filepath.Separator)) {
		return "", errors.New("object path escapes root")
	}
	return path, nil
}
