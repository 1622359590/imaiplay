package service

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/storage"
)

func TestResourceServicePlatformUploadListAndCover(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	root := t.TempDir()
	local, err := storage.NewLocal(storage.LocalConfig{
		Root: root, URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	service := NewResourceService(repository.NewResourceRepository(database), local)
	superadmin := courseContext("root", "", "superadmin")
	png := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 8)...,
	)
	cover, err := service.UploadPlatform(
		superadmin, "cover.png", bytes.NewReader(png), int64(len(png)),
	)
	if err != nil {
		t.Fatalf("UploadPlatform() error = %v", err)
	}
	if cover.TenantID != "" || cover.ResourceType != "image" ||
		cover.URL != "/api/v1/platform-covers/"+cover.ID {
		t.Fatalf("UploadPlatform() = %#v", cover)
	}
	files := regularFiles(t, root)
	if len(files) != 1 ||
		filepath.ToSlash(files[0]) != filepath.ToSlash(
			filepath.Join(root, "platform/images", cover.ID+".png"),
		) {
		t.Fatalf("platform files = %#v", files)
	}

	items, total, err := service.ListPlatform(superadmin, 0, 20)
	if err != nil || total != 1 || len(items) != 1 ||
		items[0].URL != cover.URL {
		t.Fatalf("ListPlatform() = %#v, %d, %v", items, total, err)
	}
	if _, _, err := service.ListPlatform(
		courseContext("admin", "tenant-a", "tenant_admin"), 0, 20,
	); errorCode(err) != 40300 {
		t.Fatalf("ListPlatform(tenant admin) error = %#v", err)
	}
	if _, err := service.UploadPlatform(
		courseContext("admin", "tenant-a", "tenant_admin"),
		"cover.png", bytes.NewReader(png), int64(len(png)),
	); errorCode(err) != 40300 {
		t.Fatalf("UploadPlatform(tenant admin) error = %#v", err)
	}

	body, contentType, name, err := service.OpenPlatformCover(
		context.Background(), cover.ID,
	)
	if err != nil {
		t.Fatalf("OpenPlatformCover() error = %v", err)
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil || !bytes.Equal(got, png) ||
		contentType != "image/png" || name != "cover.png" {
		t.Fatalf(
			"OpenPlatformCover() = %q, %q, %q, %v",
			got, contentType, name, err,
		)
	}
}

func TestResourceServiceOfficialResourceAccess(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	local, err := storage.NewLocal(storage.LocalConfig{
		Root: t.TempDir(), URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	service := NewResourceService(repository.NewResourceRepository(database), local)
	video := []byte{
		0, 0, 0, 24, 'f', 't', 'y', 'p',
		'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm',
	}
	resource, err := service.UploadPlatform(
		courseContext("root", "", "superadmin"),
		"lesson.mp4", bytes.NewReader(video), int64(len(video)),
	)
	if err != nil {
		t.Fatalf("UploadPlatform() error = %v", err)
	}
	course := &domain.Course{
		Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{CourseID: course.ID, Title: "Chapter"}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
		ResourceID: &resource.ID,
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	for tenantID, enabled := range map[string]bool{
		"tenant-enabled":  true,
		"tenant-disabled": false,
	} {
		if err := database.Create(&domain.TenantOfficialCourse{
			TenantID: tenantID, CourseID: course.ID, Enabled: enabled,
		}).Error; err != nil {
			t.Fatalf("set official course state: %v", err)
		}
	}
	if err := database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-enabled"},
		CourseID:  course.ID, UserID: "learner-enrolled", Status: 1,
	}).Error; err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	allowed := []context.Context{
		courseContext("root", "", "superadmin"),
		courseContext("admin", "tenant-enabled", "tenant_admin"),
		courseContext("teacher", "tenant-enabled", "instructor"),
		courseContext("learner-enrolled", "tenant-enabled", "learner"),
	}
	for _, ctx := range allowed {
		body, _, _, err := service.Open(ctx, resource.ID)
		if err != nil {
			t.Fatalf("Open(allowed) error = %v", err)
		}
		got, readErr := io.ReadAll(body)
		_ = body.Close()
		if readErr != nil || !bytes.Equal(got, video) {
			t.Fatalf("Open(allowed) content = %q, %v", got, readErr)
		}
	}
	denied := []context.Context{
		courseContext("learner-other", "tenant-enabled", "learner"),
		courseContext("admin", "tenant-disabled", "tenant_admin"),
	}
	for _, ctx := range denied {
		if _, _, _, err := service.Open(
			ctx, resource.ID,
		); errorCode(err) != 40400 {
			t.Fatalf("Open(denied) error = %#v", err)
		}
	}
	if _, _, _, err := service.OpenPlatformCover(
		context.Background(), resource.ID,
	); errorCode(err) != 40400 {
		t.Fatalf("OpenPlatformCover(video) error = %#v", err)
	}
}

func TestResourceServicePlatformDeleteRejectsReferencedResource(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	local, err := storage.NewLocal(storage.LocalConfig{
		Root: t.TempDir(), URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	service := NewResourceService(repository.NewResourceRepository(database), local)
	superadmin := courseContext("root", "", "superadmin")
	pdf := "%PDF-1.7\n"
	resource, err := service.UploadPlatform(
		superadmin, "guide.pdf", strings.NewReader(pdf), int64(len(pdf)),
	)
	if err != nil {
		t.Fatalf("UploadPlatform() error = %v", err)
	}
	course := &domain.Course{
		Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{CourseID: course.ID, Title: "Chapter"}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	if err := database.Create(&domain.CourseLesson{
		ChapterID: chapter.ID, Title: "PDF", ContentType: "document",
		ResourceID: &resource.ID,
	}).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	if err := service.DeletePlatform(
		superadmin, resource.ID,
	); errorCode(err) != 40900 {
		t.Fatalf("DeletePlatform(referenced) error = %#v", err)
	}
}
