package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalPutURLAndDelete(t *testing.T) {
	root := t.TempDir()
	local, err := NewLocal(LocalConfig{
		Root: root,
		URL:  "http://localhost:8080/uploads/",
	})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}

	url, err := local.Put(
		context.Background(), "tenant-1/asset.pdf", strings.NewReader("%PDF"), 4,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if url != "http://localhost:8080/uploads/tenant-1/asset.pdf" {
		t.Fatalf("Put() URL = %q", url)
	}
	got, err := os.ReadFile(filepath.Join(root, "tenant-1", "asset.pdf"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "%PDF" {
		t.Fatalf("file = %q", got)
	}
	if err := local.Delete(context.Background(), "tenant-1/asset.pdf"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tenant-1", "asset.pdf")); !os.IsNotExist(err) {
		t.Fatalf("deleted file Stat() error = %v", err)
	}
}

func TestLocalRejectsUnsafeKeysAndSizeMismatch(t *testing.T) {
	root := t.TempDir()
	local, err := NewLocal(LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	for _, key := range []string{"../escape.pdf", "/absolute.pdf", "safe/../../escape.pdf"} {
		if _, err := local.Put(
			context.Background(), key, strings.NewReader("data"), 4,
		); err == nil {
			t.Fatalf("Put(%q) error = nil", key)
		}
	}
	if _, err := local.Put(
		context.Background(), "short.pdf", strings.NewReader("data"), 5,
	); err == nil {
		t.Fatal("Put(size mismatch) error = nil")
	}
	if _, err := os.Stat(filepath.Join(root, "short.pdf")); !os.IsNotExist(err) {
		t.Fatalf("partial file Stat() error = %v", err)
	}
}

func TestLocalRejectsMissingConfiguration(t *testing.T) {
	tests := []LocalConfig{
		{URL: "/uploads"},
		{Root: t.TempDir()},
	}
	for _, cfg := range tests {
		if _, err := NewLocal(cfg); err == nil {
			t.Fatalf("NewLocal(%#v) error = nil", cfg)
		}
	}
}
