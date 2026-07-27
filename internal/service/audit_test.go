package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

type auditRepositoryStub struct {
	filter repository.AuditLogFilter
	detail string
}

func (stub *auditRepositoryStub) Create(_ context.Context, log *domain.AuditLog) error {
	stub.detail = log.Detail
	return nil
}
func (stub *auditRepositoryStub) List(_ context.Context, filter repository.AuditLogFilter) ([]domain.AuditLog, int64, error) {
	stub.filter = filter
	return nil, 0, nil
}

func TestAuditTenantIsolationAndSensitiveDetail(t *testing.T) {
	repositoryStub := &auditRepositoryStub{}
	audit := NewAuditService(repositoryStub)
	ctx := usercontext.WithUser(context.Background(), "user-1", "tenant-1", "admin@example.com", "tenant_admin")
	if _, _, err := audit.List(ctx, repository.AuditLogFilter{TenantID: "tenant-2", Limit: 20}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repositoryStub.filter.TenantID != "tenant-1" {
		t.Fatalf("tenant filter = %q, want tenant-1", repositoryStub.filter.TenantID)
	}
	if err := audit.Record(ctx, domain.AuditEvent{Action: "user.create", Detail: `{"password":"hidden","name":"Alice"}`}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if repositoryStub.detail != `{"name":"Alice"}` {
		t.Fatalf("stored detail = %s", repositoryStub.detail)
	}
}

func TestAuditSuperadminCanFilterTenants(t *testing.T) {
	repositoryStub := &auditRepositoryStub{}
	audit := NewAuditService(repositoryStub)
	ctx := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	if _, _, err := audit.List(ctx, repository.AuditLogFilter{TenantID: "tenant-2", Limit: 20}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repositoryStub.filter.TenantID != "tenant-2" {
		t.Fatalf("tenant filter = %q, want tenant-2", repositoryStub.filter.TenantID)
	}
}
