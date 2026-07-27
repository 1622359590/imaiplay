package storage

import (
	"context"
	"io"
)

type Storage interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64) (string, error)
	URL(key string) string
	Delete(ctx context.Context, key string) error
}

type Config struct {
	Driver string
	Local  LocalConfig
	S3     S3Config
}

type LocalConfig struct {
	Root string
	URL  string
}

type S3Config struct{}
