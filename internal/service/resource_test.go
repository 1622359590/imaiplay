package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/storage"
)

func TestResourceServiceUploadSupportedTypesAndDelete(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		content      []byte
		resourceType string
		extension    string
	}{
		{"jpeg", "photo.jpg", []byte{0xff, 0xd8, 0xff, 0xe0, 0, 16, 'J', 'F', 'I', 'F', 0}, "image", ".jpg"},
		{"png", "photo.png", append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, make([]byte, 8)...), "image", ".png"},
		{"webp", "photo.webp", []byte("RIFF\x10\x00\x00\x00WEBPVP8 "), "image", ".webp"},
		{"mp4", "video.mp4", []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm'}, "video", ".mp4"},
		{"webm", "video.webm", []byte{0x1a, 0x45, 0xdf, 0xa3, 0x9f, 0x42, 0x86, 0x81, 0x01, 0x42, 0xf7, 0x81, 0x01, 0x42, 0xf2, 0x81, 0x04, 0x42, 0xf3, 0x81, 0x08, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm'}, "video", ".webm"},
		{"pdf", "../../guide.exe", []byte("%PDF-1.7\n"), "document", ".pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			service := newResourceService(t, root)
			ctx := courseContext("admin-1", "tenant-1", "tenant_admin")
			resource, err := service.Upload(
				ctx, tt.fileName, bytes.NewReader(tt.content), int64(len(tt.content)),
			)
			if err != nil {
				t.Fatalf("Upload() error = %v", err)
			}
			if resource.ResourceType != tt.resourceType ||
				!strings.HasSuffix(resource.URL, tt.extension) ||
				strings.Contains(resource.URL, "guide.exe") ||
				strings.Contains(resource.URL, "..") {
				t.Fatalf("Upload() resource = %#v", resource)
			}
			if err := service.Delete(ctx, resource.ID); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			if files := regularFiles(t, root); len(files) != 0 {
				t.Fatalf("files after Delete() = %#v", files)
			}
		})
	}
}

func TestResourceServiceRejectsUnsupportedOversizeAndRole(t *testing.T) {
	service := newResourceService(t, t.TempDir())
	admin := courseContext("admin-1", "tenant-1", "tenant_admin")
	learner := courseContext("learner-1", "tenant-1", "learner")
	for _, test := range []struct {
		ctx  context.Context
		name string
		data []byte
		size int64
		code int
	}{
		{admin, "malware.exe", []byte("MZ executable"), 13, 40000},
		{admin, "large.pdf", []byte("%PDF"), 500*1024*1024 + 1, 40000},
		{learner, "guide.pdf", []byte("%PDF"), 4, 40300},
	} {
		_, err := service.Upload(
			test.ctx, test.name, bytes.NewReader(test.data), test.size,
		)
		var appErr *errorsx.AppError
		if !errors.As(err, &appErr) || appErr.Code != test.code {
			t.Fatalf("Upload(%s) error = %#v", test.name, err)
		}
		if test.code == 40000 &&
			appErr.Message != "unsupported file type or size exceeds limit" {
			t.Fatalf("Upload(%s) message = %q", test.name, appErr.Message)
		}
	}
}

func TestResourceServiceRemovesStoredFileWhenDatabaseCreateFails(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	root := t.TempDir()
	local, err := storage.NewLocal(storage.LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	service := NewResourceService(repository.NewResourceRepository(database), local)
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("database.DB() error = %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	pdf := "%PDF-1.7\n"
	if _, err := service.Upload(
		courseContext("admin-1", "tenant-1", "tenant_admin"),
		"guide.pdf", strings.NewReader(pdf), int64(len(pdf)),
	); errorCode(err) != 50000 {
		t.Fatalf("Upload(database failure) error = %#v", err)
	}
	if files := regularFiles(t, root); len(files) != 0 {
		t.Fatalf("files after failed Upload() = %#v", files)
	}
}

func TestResourceServiceKeepsFileWhenDatabaseDeleteFails(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	root := t.TempDir()
	local, err := storage.NewLocal(storage.LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	repo := repository.NewResourceRepository(database)
	service := NewResourceService(repo, local)
	ctx := courseContext("admin-1", "tenant-1", "tenant_admin")
	pdf := "%PDF-1.7\n"
	resource, err := service.Upload(
		ctx, "guide.pdf", strings.NewReader(pdf), int64(len(pdf)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	service.resources = failingDeleteResourceRepository{
		ResourceRepository: repo,
	}
	if err := service.Delete(ctx, resource.ID); errorCode(err) != 50000 {
		t.Fatalf("Delete(database failure) error = %#v", err)
	}
	if files := regularFiles(t, root); len(files) != 1 {
		t.Fatalf("files after failed Delete() = %#v", files)
	}
}

func TestResourceServiceRestoresRecordWhenFileDeleteFails(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	root := t.TempDir()
	local, err := storage.NewLocal(storage.LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	repo := repository.NewResourceRepository(database)
	service := NewResourceService(repo, failingDeleteStorage{Storage: local})
	ctx := courseContext("admin-1", "tenant-1", "tenant_admin")
	pdf := "%PDF-1.7\n"
	resource, err := service.Upload(
		ctx, "guide.pdf", strings.NewReader(pdf), int64(len(pdf)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := service.Delete(ctx, resource.ID); errorCode(err) != 50000 {
		t.Fatalf("Delete(storage failure) error = %#v", err)
	}
	if files := regularFiles(t, root); len(files) != 1 {
		t.Fatalf("files after failed Delete() = %#v", files)
	}
	if _, err := repo.FindByID(ctx, resource.ID); err != nil {
		t.Fatalf("resource was not restored: %v", err)
	}
}

type failingDeleteResourceRepository struct {
	repository.ResourceRepository
}

func (failingDeleteResourceRepository) Delete(context.Context, string) error {
	return errors.New("database unavailable")
}

type failingDeleteStorage struct {
	storage.Storage
}

func (failingDeleteStorage) Delete(context.Context, string) error {
	return errors.New("storage unavailable")
}

func newResourceService(t *testing.T, root string) *ResourceService {
	t.Helper()
	database, _, _ := serviceRepositories(t)
	local, err := storage.NewLocal(storage.LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	return NewResourceService(repository.NewResourceRepository(database), local)
}

func regularFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	if err := filepath.WalkDir(root, func(
		path string, entry os.DirEntry, err error,
	) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}
	return files
}
