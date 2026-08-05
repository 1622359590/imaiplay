package repository

import (
	"context"
	"errors"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestCourseRepositoryDeleteRemovesMaterialAssociationsButKeepsResources(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseRepository(database)
	for _, fixture := range []struct {
		name, tenantID, role string
		official             bool
	}{
		{"tenant", "tenant-1", "tenant_admin", false},
		{"official", "", "superadmin", true},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			course := &domain.Course{BaseModel: domain.BaseModel{TenantID: fixture.tenantID}, Title: fixture.name, Status: 1, CreatedBy: "owner", IsOfficial: fixture.official}
			resource := &domain.Resource{BaseModel: domain.BaseModel{TenantID: fixture.tenantID}, Name: fixture.name + ".pdf", ResourceType: "attachment", URL: "/uploads/" + fixture.name + ".pdf", CreatedBy: "owner"}
			if err := database.Create(course).Error; err != nil {
				t.Fatalf("create course: %v", err)
			}
			if err := database.Create(resource).Error; err != nil {
				t.Fatalf("create resource: %v", err)
			}
			material := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: fixture.tenantID}, CourseID: course.ID, ResourceID: resource.ID, DisplayName: resource.Name, CreatedBy: "owner"}
			if err := database.Create(material).Error; err != nil {
				t.Fatalf("create material: %v", err)
			}
			ctx := usercontext.WithUser(context.Background(), "owner", fixture.tenantID, "", fixture.role)
			if err := repo.Delete(ctx, course.ID); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			var materialCount, resourceCount int64
			if err := database.Model(&domain.CourseMaterial{}).Where("id = ?", material.ID).Count(&materialCount).Error; err != nil {
				t.Fatalf("count materials: %v", err)
			}
			if err := database.Model(&domain.Resource{}).Where("id = ?", resource.ID).Count(&resourceCount).Error; err != nil {
				t.Fatalf("count resources: %v", err)
			}
			if materialCount != 0 || resourceCount != 1 {
				t.Fatalf("counts materials=%d resources=%d", materialCount, resourceCount)
			}
		})
	}
}

func TestCourseRepositoryCRUDScopeAndPublishedList(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin", "tenant-1", "", "tenant_admin")
	instructor := usercontext.WithUser(base, "author-1", "tenant-1", "", "instructor")
	otherTenant := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")

	draft := newCourse("tenant-1", "author-1", "Draft", 0)
	published := newCourse("tenant-1", "author-2", "Published", 1)
	foreign := newCourse("tenant-2", "author-1", "Foreign", 1)
	for _, course := range []*domain.Course{draft, published, foreign} {
		if err := repo.Create(base, course); err != nil {
			t.Fatalf("Create(%s) error = %v", course.Title, err)
		}
	}

	if _, err := repo.FindByID(otherTenant, draft.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, total, err := repo.FindByTenant(admin, "tenant-1", 0, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("admin FindByTenant() = %#v, %d, %v", items, total, err)
	}
	items, total, err = repo.FindByTenantAndCreator(instructor, "tenant-1", "author-1", 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != draft.ID {
		t.Fatalf("instructor FindByTenantAndCreator() = %#v, %d, %v", items, total, err)
	}
	items, total, err = repo.FindPublishedByTenant(base, "tenant-1", 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != published.ID {
		t.Fatalf("FindPublishedByTenant() = %#v, %d, %v", items, total, err)
	}

	draft.Title, draft.Status = "Updated", 1
	if err := repo.Update(admin, draft); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repo.FindByID(admin, draft.ID)
	if err != nil || found.Title != "Updated" || found.Status != 1 {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Delete(otherTenant, draft.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(admin, draft.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseRepositoryFindByTenantAndCreatorAndMutationScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseRepository(database)
	base := context.Background()
	owner := usercontext.WithUser(base, "owner", "tenant-1", "", "instructor")
	owned := newCourse("tenant-1", "owner", "Owned", 1)
	foreign := newCourse("tenant-1", "other", "Foreign", 1)
	crossTenant := newCourse("tenant-2", "owner", "Cross tenant", 1)
	for _, course := range []*domain.Course{owned, foreign, crossTenant} {
		if err := repo.Create(base, course); err != nil {
			t.Fatalf("Create(%s) error = %v", course.Title, err)
		}
	}
	items, total, err := repo.FindByTenantAndCreator(base, "tenant-1", "owner", 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != owned.ID {
		t.Fatalf("FindByTenantAndCreator() = %#v, %d, %v", items, total, err)
	}
	foreign.Title = "Changed"
	if err := repo.Update(owner, foreign); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("instructor foreign Update() error = %v", err)
	}
	if err := repo.Delete(owner, foreign.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("instructor foreign Delete() error = %v", err)
	}
}

func TestCourseRepositoryFindByTenantExcludesTenantScopedOfficialRows(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseRepository(database)
	normal := &domain.Course{
		BaseModel: domain.BaseModel{ID: "tenant-normal", TenantID: "tenant-1"},
		Title:     "Normal", CreatedBy: "admin-1",
	}
	anomalous := &domain.Course{
		BaseModel: domain.BaseModel{ID: "tenant-official-anomaly", TenantID: "tenant-1"},
		Title:     "Anomalous official", CreatedBy: "root", IsOfficial: true,
	}
	for _, course := range []*domain.Course{normal, anomalous} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
	admin := usercontext.WithUser(context.Background(), "admin-1", "tenant-1", "", "tenant_admin")
	items, total, err := repo.FindByTenant(admin, "tenant-1", 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != normal.ID {
		t.Fatalf("FindByTenant() = %#v, %d, %v", items, total, err)
	}
}

func TestCourseRepositoryDeleteCascadesContentWithinTenant(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	ctx := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	)
	course := newCourse("tenant-1", "author", "Course", 0)
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  course.ID, Title: "Chapter",
	}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	enrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  course.ID, UserID: "learner", Status: 1,
	}
	progress := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner", LessonID: lesson.ID, ProgressPercent: 50,
	}
	if err := database.Create(enrollment).Error; err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	if err := database.Create(progress).Error; err != nil {
		t.Fatalf("create progress: %v", err)
	}
	if err := NewCourseRepository(database).Delete(ctx, course.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	for name, model := range map[string]interface{}{
		"course": course, "chapter": chapter, "lesson": lesson,
		"enrollment": enrollment, "progress": progress,
	} {
		var count int64
		if err := database.Model(model).Where("id = ?", modelID(model)).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("%s count=%d error=%v", name, count, err)
		}
	}
}

func TestOfficialCourseRequiresTenantActivation(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewCourseRepository(database)
	official := newCourse("", "root", "Official", 1)
	official.IsOfficial = true
	if err := repo.Create(context.Background(), official); err != nil {
		t.Fatal(err)
	}
	if _, total, err := repo.FindPublishedByTenant(context.Background(), "tenant-a", 0, 10); err != nil || total != 0 {
		t.Fatalf("unactivated official course leaked: %d, %v", total, err)
	}
	if err := repo.ActivateOfficial(context.Background(), "tenant-a", official.ID, true); err != nil {
		t.Fatal(err)
	}
	items, total, err := repo.FindPublishedByTenant(context.Background(), "tenant-a", 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != official.ID {
		t.Fatalf("activated official course missing: %#v, %d, %v", items, total, err)
	}
	if _, total, err := repo.FindPublishedByTenant(context.Background(), "tenant-b", 0, 10); err != nil || total != 0 {
		t.Fatalf("official course leaked to another tenant: %d, %v", total, err)
	}
}

func TestOfficialCourseListIncludesTenantActivation(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewCourseRepository(database)
	base := context.Background()
	draft := newCourse("", "root", "Draft official", 0)
	draft.IsOfficial = true
	available := newCourse("", "root", "Available official", 1)
	available.IsOfficial = true
	enabled := newCourse("", "root", "Enabled official", 1)
	enabled.IsOfficial = true
	for _, course := range []*domain.Course{draft, available, enabled} {
		if err := repo.Create(base, course); err != nil {
			t.Fatalf("Create(%s) error = %v", course.Title, err)
		}
	}
	if err := repo.ActivateOfficial(base, "tenant-a", enabled.ID, true); err != nil {
		t.Fatalf("ActivateOfficial() error = %v", err)
	}

	admin := usercontext.WithUser(base, "admin", "tenant-a", "", "tenant_admin")
	items, total, err := repo.FindOfficial(admin, 0, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("tenant FindOfficial() = %#v, %d, %v", items, total, err)
	}
	states := map[string]bool{}
	for _, course := range items {
		states[course.ID] = course.Enabled
	}
	if states[available.ID] || !states[enabled.ID] {
		t.Fatalf("tenant activation states = %#v", states)
	}
	if _, found := states[draft.ID]; found {
		t.Fatalf("draft official course leaked to tenant: %#v", items)
	}

	root := usercontext.WithUser(base, "root", "", "", "superadmin")
	items, total, err = repo.FindOfficial(root, 0, 10)
	if err != nil || total != 3 || len(items) != 3 {
		t.Fatalf("superadmin FindOfficial() = %#v, %d, %v", items, total, err)
	}
}

func TestCourseRepositoryDeleteOfficialCleansTenantReferences(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	course := newCourse("", "root", "Official", 1)
	course.IsOfficial = true
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{CourseID: course.ID, Title: "Chapter"}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	activation := &domain.TenantOfficialCourse{
		TenantID: "tenant-a", CourseID: course.ID, Enabled: true,
	}
	enrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-a"},
		CourseID:  course.ID, UserID: "learner", Status: 1,
	}
	progress := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-a"},
		UserID:    "learner", LessonID: lesson.ID, ProgressPercent: 50,
	}
	for name, model := range map[string]interface{}{
		"activation": activation,
		"enrollment": enrollment,
		"progress":   progress,
	} {
		if err := database.Create(model).Error; err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	ctx := usercontext.WithUser(
		context.Background(), "root", "", "", "superadmin",
	)
	if err := NewCourseRepository(database).Delete(ctx, course.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	for name, model := range map[string]interface{}{
		"course": course, "chapter": chapter, "lesson": lesson,
		"enrollment": enrollment, "progress": progress,
	} {
		var count int64
		if err := database.Model(model).
			Where("id = ?", modelID(model)).
			Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("%s count=%d error=%v", name, count, err)
		}
	}
	var activationCount int64
	if err := database.Model(&domain.TenantOfficialCourse{}).
		Where("tenant_id = ? AND course_id = ?", "tenant-a", course.ID).
		Count(&activationCount).Error; err != nil || activationCount != 0 {
		t.Fatalf(
			"activation count=%d error=%v",
			activationCount, err,
		)
	}
}

func modelID(model interface{}) string {
	switch value := model.(type) {
	case *domain.Course:
		return value.ID
	case *domain.CourseChapter:
		return value.ID
	case *domain.CourseLesson:
		return value.ID
	case *domain.CourseEnrollment:
		return value.ID
	case *domain.LessonProgress:
		return value.ID
	default:
		return ""
	}
}

func newCourse(tenantID, creator, title string, status int) *domain.Course {
	return &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Title:     title, Status: status, CreatedBy: creator,
	}
}
