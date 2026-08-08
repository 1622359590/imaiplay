package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPortalResolveByTenantCode(t *testing.T) {
	service, tenants := newPortalTestService(t)
	tenant := &domain.Tenant{
		Code:                    "acme",
		Name:                    "Acme 学院",
		Status:                  1,
		PrimaryColor:            "#123456",
		SelectedBackgroundColor: "#FFF1F0",
		SelectedTextColor:       "#C5221F",
		SelectedIconColor:       "#8C1D18",
		WelcomeText:             "欢迎学习",
	}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}

	portal, err := service.Resolve(context.Background(), " ACME ", "play.imai.work")
	if err != nil {
		t.Fatal(err)
	}
	if portal.Code != "acme" || portal.Name != "Acme 学院" {
		t.Fatalf("portal=%#v", portal)
	}
	encoded, err := json.Marshal(portal)
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]string
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	if response["tenant_id"] != tenant.ID {
		t.Fatalf("tenant_id=%q want=%q", response["tenant_id"], tenant.ID)
	}
	if portal.DefaultPortalURL != "https://play.imai.work/t/acme" {
		t.Fatalf("default portal URL=%q", portal.DefaultPortalURL)
	}
	if portal.PrimaryColor != "#123456" || portal.WelcomeText != "欢迎学习" {
		t.Fatalf("branding=%#v", portal)
	}
	if portal.SelectedBackgroundColor != "#FFF1F0" ||
		portal.SelectedTextColor != "#C5221F" ||
		portal.SelectedIconColor != "#8C1D18" {
		t.Fatalf("selection colors=%#v", portal)
	}
}

func TestPortalResolveByCustomDomain(t *testing.T) {
	service, tenants := newPortalTestService(t)
	customDomain := "learn.acme.test"
	tenant := &domain.Tenant{
		Code:         "acme",
		Name:         "Acme 学院",
		Status:       1,
		CustomDomain: &customDomain,
	}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}

	portal, err := service.Resolve(
		context.Background(),
		"",
		"LEARN.ACME.TEST:443",
	)
	if err != nil {
		t.Fatal(err)
	}
	if portal.Code != tenant.Code {
		t.Fatalf("portal code=%q want=%q", portal.Code, tenant.Code)
	}
	if portal.CustomDomainURL != "https://learn.acme.test" {
		t.Fatalf("custom domain URL=%q", portal.CustomDomainURL)
	}
}

func TestPortalResolveCustomDomainOverridesConflictingTenantCode(t *testing.T) {
	portalService, tenants := newPortalTestService(t)
	customDomain := "learn.acme.test"
	acme := &domain.Tenant{
		Code: "acme", Name: "Acme 学院", Status: 1, CustomDomain: &customDomain,
	}
	bravo := &domain.Tenant{Code: "bravo", Name: "Bravo 学院", Status: 1}
	for _, tenant := range []*domain.Tenant{acme, bravo} {
		if err := tenants.Create(context.Background(), tenant); err != nil {
			t.Fatal(err)
		}
	}

	portal, err := portalService.Resolve(
		context.Background(),
		"bravo",
		"LEARN.ACME.TEST:443",
	)
	if err != nil {
		t.Fatal(err)
	}
	if portal.Code != "acme" {
		t.Fatalf("portal code=%q want=acme", portal.Code)
	}

	_, err = portalService.Resolve(
		context.Background(),
		"bravo",
		"unverified.example.test",
	)
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40400 {
		t.Fatalf("error=%v want unverified custom host not found", err)
	}
}

func TestPortalResolveByAuthenticatedTenantID(t *testing.T) {
	portalService, tenants := newPortalTestService(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme 学院", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}

	portal, err := portalService.ResolveByTenantID(context.Background(), tenant.ID)
	if err != nil {
		t.Fatal(err)
	}
	if portal.TenantID != tenant.ID || portal.Code != tenant.Code {
		t.Fatalf("portal=%#v", portal)
	}
}

func TestPortalResolveRejectsUnavailableTenant(t *testing.T) {
	service, tenants := newPortalTestService(t)
	tenant := &domain.Tenant{
		Code:            "suspended",
		Name:            "Suspended",
		Status:          1,
		LifecycleStatus: "suspended",
	}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}

	_, err := service.Resolve(context.Background(), tenant.Code, "play.imai.work")
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40300 {
		t.Fatalf("error=%v want forbidden", err)
	}
}

func TestPortalResolveReturnsNotFound(t *testing.T) {
	service, _ := newPortalTestService(t)
	_, err := service.Resolve(context.Background(), "missing", "play.imai.work")
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40400 {
		t.Fatalf("error=%v want not found", err)
	}
}

func newPortalTestService(
	t *testing.T,
) (*PortalService, repository.TenantRepository) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	tenants := repository.NewTenantRepository(database)
	return NewPortalService(tenants, "play.imai.work"), tenants
}
