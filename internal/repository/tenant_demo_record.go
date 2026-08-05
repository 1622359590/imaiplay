package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

const (
	DemoRecordCourse        = "course"
	DemoRecordCourseChapter = "course_chapter"
	DemoRecordCourseLesson  = "course_lesson"
	DemoRecordUser          = "user"
	DemoRecordResource      = "resource"
)

type TenantDemoRecordRepository interface {
	RegisterBatch(ctx context.Context, records []domain.TenantDemoRecord) error
	HasRecords(ctx context.Context, tenantID string) (bool, error)
	ListByTenant(ctx context.Context, tenantID string) ([]domain.TenantDemoRecord, error)
	DeleteBatch(ctx context.Context, tenantID, batchID string) error
}
