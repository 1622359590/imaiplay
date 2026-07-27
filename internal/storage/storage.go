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
	Driver string      `json:"driver"`
	Local  LocalConfig `json:"local"`
	S3     S3Config    `json:"s3"`
}

type LocalConfig struct {
	Root string `json:"root"`
	URL  string `json:"url"`
}

type S3Config struct {
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key,omitempty"`
	Region    string `json:"region"`
	Prefix    string `json:"prefix"`
}
