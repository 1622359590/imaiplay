package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
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

func TestResourceServiceFileScopesTenantAndRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	service := newResourceService(t, root)
	content := []byte("%PDF-1.7\nprivate document")
	resource, err := service.Upload(
		courseContext("admin-1", "tenant-1", "tenant_admin"),
		"guide.pdf", bytes.NewReader(content), int64(len(content)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	path, contentType, name, err := service.File(
		courseContext("admin-1", "tenant-1", "tenant_admin"), resource.ID, root,
	)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("File() content = %q, error = %v", got, err)
	}
	if contentType != "application/pdf" || name != "guide.pdf" {
		t.Fatalf("File() metadata = %q, %q", contentType, name)
	}
	if _, _, _, err := service.File(
		courseContext("admin-2", "tenant-2", "tenant_admin"), resource.ID, root,
	); errorCode(err) != 40400 {
		t.Fatalf("cross-tenant File() error = %#v", err)
	}

	database, _, _ := serviceRepositories(t)
	malicious := &domain.Resource{
		BaseModel:    domain.BaseModel{TenantID: "tenant-1"},
		Name:         "escape.pdf",
		ResourceType: "document",
		URL:          "/uploads/../escape.pdf",
		CreatedBy:    "admin-1",
	}
	if err := database.Create(malicious).Error; err != nil {
		t.Fatalf("create malicious resource: %v", err)
	}
	// The malicious record is in a separate test database and only verifies
	// that the service rejects traversal keys before touching the filesystem.
	maliciousService := NewResourceService(repository.NewResourceRepository(database), service.storage)
	if _, _, _, err := maliciousService.File(
		courseContext("admin-1", "tenant-1", "tenant_admin"), malicious.ID, root,
	); errorCode(err) != 40400 {
		t.Fatalf("traversal File() error = %#v", err)
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
		{admin, "large.pdf", []byte("%PDF"), 1024*1024*1024 + 1, 40000},
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

func TestResourceServiceInstructorCanUploadAndListButCannotDelete(t *testing.T) {
	service := newResourceService(t, t.TempDir())
	instructor := courseContext("instructor-1", "tenant-1", "instructor")
	body := []byte("%PDF-1.7\n")
	resource, err := service.UploadAttachment(
		instructor, "guide.pdf", bytes.NewReader(body), int64(len(body)),
	)
	if err != nil || resource.CreatedBy != "instructor-1" || resource.TenantID != "tenant-1" {
		t.Fatalf("UploadAttachment() = %#v, %v", resource, err)
	}
	items, total, err := service.List(instructor, 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != resource.ID {
		t.Fatalf("List() = %#v, %d, %v", items, total, err)
	}
	if err := service.Delete(instructor, resource.ID); errorCode(err) != 40300 {
		t.Fatalf("Delete() error = %#v", err)
	}
}

func TestResourceUploadLimitIsOneGiB(t *testing.T) {
	const oneGiB int64 = 1024 * 1024 * 1024
	if maxResourceSize != oneGiB {
		t.Fatalf("maxResourceSize = %d, want %d", maxResourceSize, oneGiB)
	}
}

func TestResourceServiceUploadAttachmentValidatesSignatureAndExtension(t *testing.T) {
	tests := []struct {
		name   string
		body   []byte
		suffix string
	}{
		{"guide.pdf", []byte("%PDF-1.7\n"), ".pdf"},
		{"guide.doc", append([]byte{0xd0, 0xcf, 0x11, 0xe0}, make([]byte, 508)...), ".doc"},
		{"sheet.xlsx", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".xlsx"},
		{"slides.pptx", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".pptx"},
		{"bundle.zip", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".zip"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newResourceService(t, t.TempDir())
			resource, err := service.UploadAttachment(
				courseContext("admin", "tenant-1", "tenant_admin"),
				test.name, bytes.NewReader(test.body), int64(len(test.body)),
			)
			if err != nil {
				t.Fatalf("UploadAttachment() error = %v", err)
			}
			if resource.ResourceType != "attachment" || !strings.HasSuffix(resource.URL, test.suffix) {
				t.Fatalf("UploadAttachment() = %#v", resource)
			}
		})
	}

	service := newResourceService(t, t.TempDir())
	for _, test := range []struct {
		name string
		body []byte
		size int64
	}{
		{"malware.exe", []byte("MZ"), 2},
		{"guide.pdf", []byte{'P', 'K', 3, 4}, 4},
		{"guide.zip", []byte{0xd0, 0xcf, 0x11, 0xe0}, 4},
		{"large.pdf", []byte("%PDF"), 200*1024*1024 + 1},
	} {
		if _, err := service.UploadAttachment(
			courseContext("admin", "tenant-1", "tenant_admin"),
			test.name, bytes.NewReader(test.body), test.size,
		); errorCode(err) != 40000 {
			t.Fatalf("UploadAttachment(%s) error = %#v", test.name, err)
		}
	}
}

func TestResourceServiceUploadPlatformAttachmentUsesPlatformAttachmentPath(t *testing.T) {
	root := t.TempDir()
	service := newResourceService(t, root)
	body := []byte("%PDF-1.7\n")
	resource, err := service.UploadPlatformAttachment(
		courseContext("root", "", "superadmin"),
		"guide.pdf", bytes.NewReader(body), int64(len(body)),
	)
	if err != nil {
		t.Fatalf("UploadPlatformAttachment() error = %v", err)
	}
	files := regularFiles(t, root)
	if resource.ResourceType != "attachment" || len(files) != 1 || !strings.Contains(filepath.ToSlash(files[0]), "/platform/attachments/") {
		t.Fatalf("UploadPlatformAttachment() = %#v", resource)
	}
}

func TestResourceServiceRejectsDeletingCourseMaterialResource(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	root := t.TempDir()
	local, err := storage.NewLocal(storage.LocalConfig{Root: root, URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	service := NewResourceService(repository.NewResourceRepository(database), local)
	ctx := courseContext("admin", "tenant-1", "tenant_admin")
	body := []byte("%PDF-1.7\n")
	resource, err := service.UploadAttachment(ctx, "guide.pdf", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("UploadAttachment() error = %v", err)
	}
	material := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: "course-1", ResourceID: resource.ID, DisplayName: "guide.pdf", CreatedBy: "admin"}
	if err := database.Create(material).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	if err := service.Delete(ctx, resource.ID); errorCode(err) != 40900 {
		t.Fatalf("Delete(referenced resource) error = %#v", err)
	}
	if files := regularFiles(t, root); len(files) != 1 {
		t.Fatalf("files after rejected delete = %#v", files)
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
