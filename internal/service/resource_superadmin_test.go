package service

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
)

func TestResourceServiceListAllRequiresSuperadmin(t *testing.T) {
	service := newResourceService(t, t.TempDir())
	for _, resource := range []*domain.Resource{
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1"},
			Name:      "one.pdf", ResourceType: "document", URL: "/uploads/one.pdf",
			CreatedBy: "admin-1",
		},
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-2"},
			Name:      "two.mp4", ResourceType: "video", URL: "/uploads/two.mp4",
			CreatedBy: "admin-2",
		},
	} {
		if err := service.resources.Create(context.Background(), resource); err != nil {
			t.Fatalf("Create(%s) error = %v", resource.Name, err)
		}
	}
	items, total, err := service.ListAll(
		courseContext("root", "", "superadmin"), 0, 20,
	)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("ListAll(superadmin) = %#v, %d, %v", items, total, err)
	}
	if _, _, err := service.ListAll(
		courseContext("admin-1", "tenant-1", "tenant_admin"), 0, 20,
	); errorCode(err) != 40300 {
		t.Fatalf("ListAll(tenant_admin) error = %#v", err)
	}
}
