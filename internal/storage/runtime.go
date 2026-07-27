package storage

import (
	"context"
	"errors"
	"io"
	"sync"
)

type Runtime struct {
	mu      sync.RWMutex
	current Storage
	local   *Local
	store   *ConfigStore
}

func NewRuntime(local *Local, store *ConfigStore, driver string) (*Runtime, error) {
	runtime := &Runtime{current: local, local: local, store: store}
	config := store.internal()
	if driver == "s3" || config.Driver == "s3" {
		if config.S3.SecretKey == "" {
			return runtime, nil
		}
		s3, err := NewS3(config.S3)
		if err != nil {
			return nil, err
		}
		runtime.current = s3
	}
	return runtime, nil
}
func (runtime *Runtime) Put(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.current.Put(ctx, key, reader, size)
}
func (runtime *Runtime) URL(key string) string {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.current.URL(key)
}
func (runtime *Runtime) Delete(ctx context.Context, key string) error {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.current.Delete(ctx, key)
}
func (runtime *Runtime) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	runtime.mu.RLock()
	current := runtime.current
	runtime.mu.RUnlock()
	if readable, ok := current.(interface {
		Get(context.Context, string) (io.ReadCloser, error)
	}); ok {
		return readable.Get(ctx, key)
	}
	return nil, errors.New("storage backend does not support reading")
}
func (runtime *Runtime) Config() Config { return runtime.store.Get() }
func (runtime *Runtime) Save(config Config) error {
	var next Storage = runtime.local
	if config.Driver == "s3" {
		s3, err := NewS3(config.S3)
		if err != nil {
			return err
		}
		next = s3
	}
	if err := runtime.store.Save(config); err != nil {
		return err
	}
	runtime.mu.Lock()
	runtime.current = next
	runtime.mu.Unlock()
	return nil
}
func (runtime *Runtime) Test(ctx context.Context, config Config) error {
	if config.Driver != "s3" {
		return nil
	}
	s3, err := NewS3(config.S3)
	if err != nil {
		return err
	}
	return s3.Test(ctx)
}
