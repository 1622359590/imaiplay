package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	root string
	url  string
}

func NewLocal(cfg LocalConfig) (*Local, error) {
	if strings.TrimSpace(cfg.Root) == "" || strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("local storage root and URL are required")
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	return &Local{root: root, url: strings.TrimRight(cfg.URL, "/")}, nil
}

func (local *Local) Put(
	ctx context.Context, key string, reader io.Reader, size int64,
) (url string, err error) {
	path, err := local.path(key)
	if err != nil {
		return "", err
	}
	if size < 0 {
		return "", errors.New("file size cannot be negative")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create storage directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("create storage file: %w", err)
	}
	defer func() {
		closeErr := file.Close()
		if err != nil {
			_ = os.Remove(path)
		} else if closeErr != nil {
			err = closeErr
		}
	}()
	written, err := io.Copy(file, io.LimitReader(reader, size+1))
	if err != nil {
		return "", fmt.Errorf("write storage file: %w", err)
	}
	if written != size {
		return "", fmt.Errorf("file size mismatch: got %d, want %d", written, size)
	}
	return local.URL(key), nil
}

func (local *Local) URL(key string) string {
	if key == "" {
		return local.url
	}
	return local.url + "/" + filepath.ToSlash(filepath.Clean(key))
}

func (local *Local) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := local.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete storage file: %w", err)
	}
	return nil
}

func (local *Local) path(key string) (string, error) {
	if key == "" || filepath.IsAbs(key) {
		return "", errors.New("invalid storage key")
	}
	clean := filepath.Clean(key)
	if clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid storage key")
	}
	path := filepath.Join(local.root, clean)
	relative, err := filepath.Rel(local.root, path)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid storage key")
	}
	return path, nil
}
