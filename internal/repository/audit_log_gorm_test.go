package repository

import (
	"context"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestAuditLogRepositoryListFilters(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewAuditLogRepository(database)
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Second)
	logs := []*domain.AuditLog{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", CreatedAt: base.Add(-3 * time.Hour)}, UserID: "user-1", Action: "course.create", ResourceType: "course"},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", CreatedAt: base.Add(-2 * time.Hour)}, UserID: "user-2", Action: "course.update", ResourceType: "course"},
		{BaseModel: domain.BaseModel{TenantID: "tenant-2", CreatedAt: base.Add(-time.Hour)}, UserID: "user-1", Action: "course.create", ResourceType: "course"},
	}
	for _, log := range logs {
		if err := repo.Create(ctx, log); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		filter    AuditLogFilter
		wantTotal int64
		wantID    string
	}{
		{name: "tenant", filter: AuditLogFilter{TenantID: "tenant-1", Limit: 10}, wantTotal: 2, wantID: logs[1].ID},
		{name: "user", filter: AuditLogFilter{UserID: "user-2", Limit: 10}, wantTotal: 1, wantID: logs[1].ID},
		{name: "action", filter: AuditLogFilter{Action: "course.create", Limit: 10}, wantTotal: 2, wantID: logs[2].ID},
		{name: "time range", filter: AuditLogFilter{From: ptrTime(base.Add(-150 * time.Minute)), To: ptrTime(base.Add(-30 * time.Minute)), Limit: 10}, wantTotal: 2, wantID: logs[2].ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.List(ctx, tt.filter)
			if err != nil || total != tt.wantTotal || len(items) != int(tt.wantTotal) || items[0].ID != tt.wantID {
				t.Fatalf("List() = %#v, %d, %v; want %s", items, total, err, tt.wantID)
			}
		})
	}

	items, total, err := repo.List(ctx, AuditLogFilter{TenantID: "tenant-1", Offset: 1, Limit: 1})
	if err != nil || total != 2 || len(items) != 1 || items[0].ID != logs[0].ID {
		t.Fatalf("List(pagination) = %#v, %d, %v", items, total, err)
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
