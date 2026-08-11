package service

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/1622359590/imaiplay/internal/repository"
)

type domainResolverStub struct {
	cname string
	ips   []net.IP
}

func (stub domainResolverStub) LookupCNAME(string) (string, error) {
	if stub.cname == "" {
		return "", errors.New("no cname")
	}
	return stub.cname, nil
}

func (stub domainResolverStub) LookupIP(string) ([]net.IP, error) {
	if len(stub.ips) == 0 {
		return nil, errors.New("no address")
	}
	return stub.ips, nil
}

type domainPanelStub struct {
	mu          sync.Mutex
	calls       []string
	failAt      string
	failDeletes int
	sslStatus   string
}

func (stub *domainPanelStub) record(name string) error {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.calls = append(stub.calls, name)
	if name == "delete" && stub.failDeletes > 0 {
		stub.failDeletes--
		return errors.New("delete failed")
	}
	if stub.failAt == name {
		return errors.New(name + " failed")
	}
	return nil
}

func (stub *domainPanelStub) AddSite(string) (int, error) {
	if err := stub.record("site"); err != nil {
		return 0, err
	}
	return 42, nil
}

func (stub *domainPanelStub) AddReverseProxy(string, string, string) error {
	return stub.record("proxy")
}

func (stub *domainPanelStub) ApplyLetsEncrypt(string) error {
	return stub.record("ssl")
}

func (stub *domainPanelStub) EnableHTTPSRedirect(string) error {
	return stub.record("https_redirect")
}

func (stub *domainPanelStub) AddNginxSnippet(string, string) error {
	return stub.record("snippet")
}

func (stub *domainPanelStub) DeleteSite(string) error {
	return stub.record("delete")
}

func (stub *domainPanelStub) GetSiteInfo(string) (map[string]interface{}, error) {
	if err := stub.record("info"); err != nil {
		return nil, err
	}
	return map[string]interface{}{"ssl_status": stub.sslStatus}, nil
}

func (stub *domainPanelStub) callNames() []string {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return append([]string(nil), stub.calls...)
}

type domainAuditStub struct {
	mu     sync.Mutex
	events []domain.AuditEvent
}

func (stub *domainAuditStub) Record(_ context.Context, event domain.AuditEvent) error {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.events = append(stub.events, event)
	return nil
}

func TestVerifyDomainAcceptsCNAMEOrResolvedServerIP(t *testing.T) {
	t.Run("cname", func(t *testing.T) {
		ok, err := verifyDomainWithResolver(
			"academy.example.com", "120.25.77.204",
			domainResolverStub{
				cname: "play.imai.work.",
				ips:   []net.IP{net.ParseIP("120.25.77.204")},
			},
		)
		if err != nil || !ok {
			t.Fatalf("VerifyDomain() = %v, %v", ok, err)
		}
	})

	t.Run("resolved address", func(t *testing.T) {
		ok, err := verifyDomainWithResolver(
			"academy.example.com", "120.25.77.204",
			domainResolverStub{ips: []net.IP{net.ParseIP("120.25.77.204")}},
		)
		if err != nil || !ok {
			t.Fatalf("VerifyDomain() = %v, %v", ok, err)
		}
	})

	t.Run("wrong address", func(t *testing.T) {
		ok, err := verifyDomainWithResolver(
			"academy.example.com", "120.25.77.204",
			domainResolverStub{ips: []net.IP{net.ParseIP("203.0.113.8")}},
		)
		if err == nil || ok {
			t.Fatalf("VerifyDomain() = %v, %v, want rejection", ok, err)
		}
	})
}

func TestDomainBindVerifyRejectsReservedInvalidAndDuplicateDomains(t *testing.T) {
	tenants := domainBindTenants(t)
	current := &domain.Tenant{Code: "current", Name: "Current", Status: 1}
	otherDomain := "used.example.com"
	other := &domain.Tenant{Code: "other", Name: "Other", Status: 1, CustomDomain: &otherDomain}
	if err := tenants.Create(context.Background(), current); err != nil {
		t.Fatal(err)
	}
	if err := tenants.Create(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "admin", current.ID, "admin@example.com", "tenant_admin")
	svc := NewDomainBindService(tenants, &domainPanelStub{}, domainResolverStub{
		ips: []net.IP{net.ParseIP("120.25.77.204")},
	}, nil, DomainBindConfig{
		ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
		CNAMETarget: "play.imai.work", ProxyTarget: "http://127.0.0.1:18080",
	})

	for _, value := range []string{"play.imai.work", "*.example.com", "https://bad.example.com", "not-a-domain", "used.example.com"} {
		if _, err := svc.Verify(ctx, value); err == nil {
			t.Fatalf("Verify(%q) error = nil", value)
		}
	}
}

func TestDomainBindFlowTransitionsToReadyAndPersistsDomain(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := middleware.WithRequestID(
		usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin"),
		"domain-bind-request",
	)
	panel := &domainPanelStub{sslStatus: "ready"}
	audit := &domainAuditStub{}
	svc := NewDomainBindService(tenants, panel, domainResolverStub{
		ips: []net.IP{net.ParseIP("120.25.77.204")},
	}, audit, DomainBindConfig{
		ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
		CNAMETarget: "play.imai.work", ProxyTarget: "http://127.0.0.1:18080",
		PollInterval: time.Millisecond, MaxPolls: 2,
	})

	verified, err := svc.Verify(ctx, "academy.acme.com")
	if err != nil || verified.State != DomainStateVerified {
		t.Fatalf("Verify() = %#v, %v", verified, err)
	}
	started, err := svc.Bind(ctx, "academy.acme.com")
	if err != nil || started.State != DomainStateCreatingSite {
		t.Fatalf("Bind() = %#v, %v", started, err)
	}
	ready := waitDomainState(t, svc, ctx, DomainStateReady)
	if ready.Domain != "academy.acme.com" {
		t.Fatalf("ready domain = %q", ready.Domain)
	}
	if want := []string{
		DomainStatePendingVerification,
		DomainStateVerified,
		DomainStateCreatingSite,
		DomainStateConfiguring,
		DomainStateReady,
	}; !reflect.DeepEqual(ready.History, want) {
		t.Fatalf("history = %#v, want %#v", ready.History, want)
	}
	if want := []string{"site", "proxy", "snippet", "ssl", "info", "https_redirect"}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel calls = %#v, want %#v", panel.callNames(), want)
	}
	stored, err := tenants.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.CustomDomain == nil || *stored.CustomDomain != "academy.acme.com" {
		t.Fatalf("stored tenant = %#v, %v", stored, err)
	}
	if len(audit.events) != 1 || audit.events[0].Action != "domain.bind" {
		t.Fatalf("audit events = %#v", audit.events)
	}
	if audit.events[0].RequestID != "domain-bind-request" {
		t.Fatalf("audit request ID = %q", audit.events[0].RequestID)
	}
}

func TestDomainBindHTTPSRedirectFailureRollsBackWithoutPersistingDomain(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "https-rollback", Name: "HTTPS Rollback", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(
		context.Background(),
		"admin",
		tenant.ID,
		"admin@example.com",
		"tenant_admin",
	)
	panel := &domainPanelStub{failAt: "https_redirect", sslStatus: "ready"}
	audit := &domainAuditStub{}
	svc := NewDomainBindService(
		tenants,
		panel,
		domainResolverStub{ips: []net.IP{net.ParseIP("120.25.77.204")}},
		audit,
		DomainBindConfig{
			ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
			CNAMETarget: "play.imai.work", ProxyTarget: "http://127.0.0.1:18080",
			PollInterval: time.Millisecond, MaxPolls: 1,
		},
	)

	if _, err := svc.Verify(ctx, "secure.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Bind(ctx, "secure.example.com"); err != nil {
		t.Fatal(err)
	}
	failed := waitDomainState(t, svc, ctx, DomainStateSetupFailed)
	if !strings.Contains(failed.Message, "开启强制 HTTPS 失败") {
		t.Fatalf("failure message = %q", failed.Message)
	}
	if want := []string{
		"site", "proxy", "snippet", "ssl", "info", "https_redirect", "delete",
	}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel calls = %#v, want %#v", panel.callNames(), want)
	}
	stored, err := tenants.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.CustomDomain != nil {
		t.Fatalf("stored tenant = %#v, %v", stored, err)
	}
	for _, event := range audit.events {
		if event.Action == "domain.bind" {
			t.Fatalf("unexpected success audit event = %#v", event)
		}
	}
}

func TestSuperadminCanBindDomainForTargetTenant(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "platform-target", Name: "Platform Target", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "platform-admin", "", "platform@example.com", "superadmin")
	audit := &domainAuditStub{}
	svc := NewDomainBindService(tenants, &domainPanelStub{sslStatus: "ready"}, domainResolverStub{
		ips: []net.IP{net.ParseIP("120.25.77.204")},
	}, audit, DomainBindConfig{
		ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
		CNAMETarget: "play.imai.work", ProxyTarget: "http://127.0.0.1:18080",
		PollInterval: time.Millisecond, MaxPolls: 2,
	})

	if _, err := svc.VerifyForTenant(ctx, tenant.ID, "target.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindForTenant(ctx, tenant.ID, "target.example.com"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		status, err := svc.StatusForTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == DomainStateReady {
			if len(audit.events) != 1 || audit.events[0].UserRole != "superadmin" || audit.events[0].TenantID != tenant.ID {
				t.Fatalf("audit event = %#v", audit.events)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("target tenant domain did not become ready")
}

func TestDomainBindFailureRollsBackCreatedSite(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "rollback", Name: "Rollback", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	panel := &domainPanelStub{failAt: "snippet", sslStatus: "ready"}
	svc := NewDomainBindService(tenants, panel, domainResolverStub{
		ips: []net.IP{net.ParseIP("120.25.77.204")},
	}, nil, DomainBindConfig{
		ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
		CNAMETarget: "play.imai.work", ProxyTarget: "http://127.0.0.1:18080",
		PollInterval: time.Millisecond, MaxPolls: 1,
	})
	if _, err := svc.Verify(ctx, "rollback.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Bind(ctx, "rollback.example.com"); err != nil {
		t.Fatal(err)
	}
	failed := waitDomainState(t, svc, ctx, DomainStateSetupFailed)
	if failed.Message == "" {
		t.Fatal("failure reason is empty")
	}
	if strings.Contains(failed.Message, "snippet failed") {
		t.Fatalf("failure leaked BaoTa error details: %q", failed.Message)
	}
	if want := []string{"site", "proxy", "snippet", "delete"}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel calls = %#v, want %#v", panel.callNames(), want)
	}
	stored, err := tenants.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.CustomDomain != nil {
		t.Fatalf("stored tenant = %#v, %v", stored, err)
	}
	cleaned, err := svc.Unbind(ctx)
	if err != nil || cleaned.State != DomainStateNone {
		t.Fatalf("cleanup Unbind() = %#v, %v", cleaned, err)
	}
	if want := []string{"site", "proxy", "snippet", "delete"}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel cleanup calls = %#v, want %#v", panel.callNames(), want)
	}
}

func TestDomainBindRequiresFailedRollbackCleanupBeforeNewVerification(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "cleanup", Name: "Cleanup", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(
		context.Background(),
		"admin",
		tenant.ID,
		"admin@example.com",
		"tenant_admin",
	)
	panel := &domainPanelStub{
		failAt: "snippet", failDeletes: 1, sslStatus: "ready",
	}
	svc := NewDomainBindService(
		tenants,
		panel,
		domainResolverStub{ips: []net.IP{net.ParseIP("120.25.77.204")}},
		nil,
		DomainBindConfig{
			ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
			CNAMETarget: "play.imai.work", PollInterval: time.Millisecond, MaxPolls: 1,
		},
	)
	if _, err := svc.Verify(ctx, "failed.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Bind(ctx, "failed.example.com"); err != nil {
		t.Fatal(err)
	}
	waitDomainState(t, svc, ctx, DomainStateSetupFailed)

	if _, err := svc.Verify(ctx, "next.example.com"); err == nil {
		t.Fatal("Verify() error = nil while failed site still needs cleanup")
	}
	current, err := svc.Status(ctx)
	if err != nil || current.Domain != "failed.example.com" ||
		current.State != DomainStateSetupFailed {
		t.Fatalf("status after blocked verification = %#v, %v", current, err)
	}
	cleaned, err := svc.Unbind(ctx)
	if err != nil || cleaned.State != DomainStateNone {
		t.Fatalf("Unbind() = %#v, %v", cleaned, err)
	}
	if want := []string{
		"site", "proxy", "snippet", "delete", "delete",
	}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel calls = %#v, want %#v", panel.callNames(), want)
	}
}

func TestDomainBindReservesVerifiedDomainAcrossTenants(t *testing.T) {
	tenants := domainBindTenants(t)
	first := &domain.Tenant{Code: "first", Name: "First", Status: 1}
	second := &domain.Tenant{Code: "second", Name: "Second", Status: 1}
	if err := tenants.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := tenants.Create(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	panel := &domainPanelStub{sslStatus: "ready"}
	svc := NewDomainBindService(
		tenants,
		panel,
		domainResolverStub{ips: []net.IP{net.ParseIP("120.25.77.204")}},
		nil,
		DomainBindConfig{
			ExpectedIP: "120.25.77.204", ReservedDomain: "play.imai.work",
			CNAMETarget: "play.imai.work", PollInterval: time.Millisecond, MaxPolls: 2,
		},
	)
	firstContext := usercontext.WithUser(
		context.Background(), "first-admin", first.ID, "first@example.com", "tenant_admin",
	)
	secondContext := usercontext.WithUser(
		context.Background(), "second-admin", second.ID, "second@example.com", "tenant_admin",
	)
	for _, item := range []context.Context{firstContext, secondContext} {
		if _, err := svc.Verify(item, "shared.example.com"); err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
	}
	if _, err := svc.Bind(firstContext, "shared.example.com"); err != nil {
		t.Fatalf("first Bind() error = %v", err)
	}
	if _, err := svc.Bind(secondContext, "shared.example.com"); err == nil {
		t.Fatal("second Bind() error = nil, want domain reservation conflict")
	}
	waitDomainState(t, svc, firstContext, DomainStateReady)
	siteCalls := 0
	for _, call := range panel.callNames() {
		if call == "site" {
			siteCalls++
		}
	}
	if siteCalls != 1 {
		t.Fatalf("site calls = %d, want 1", siteCalls)
	}
}

func TestDomainBindStatusIncludesTenantPortalMetadataForCachedAndLoadedStates(t *testing.T) {
	tenants := domainBindTenants(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	svc := NewDomainBindService(
		tenants,
		nil,
		domainResolverStub{},
		nil,
		DomainBindConfig{ReservedDomain: "play.imai.work"},
	)

	loaded, err := svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TenantCode != "acme" || loaded.DefaultPortalURL != "https://play.imai.work/t/acme" {
		t.Fatalf("loaded status = %#v", loaded)
	}

	cached, err := svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cached.TenantCode != "acme" || cached.DefaultPortalURL != "https://play.imai.work/t/acme" {
		t.Fatalf("cached status = %#v", cached)
	}
}

func TestDomainBindStatusSurvivesServiceRestart(t *testing.T) {
	database, tenants, _ := serviceRepositories(t)
	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDatabase.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	tenant := &domain.Tenant{Code: "restart", Name: "Restart", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	jobs := repository.NewDomainBindJobRepository(database)
	config := DomainBindConfig{ExpectedIP: "120.25.77.204", CNAMETarget: "play.imai.work"}
	resolver := domainResolverStub{ips: []net.IP{net.ParseIP("120.25.77.204")}}
	first := NewDomainBindService(tenants, &domainPanelStub{}, resolver, nil, config, jobs)
	verified, err := first.Verify(ctx, "restart.example.com")
	if err != nil || verified.State != DomainStateVerified {
		t.Fatalf("Verify() = %#v, %v", verified, err)
	}

	restarted := NewDomainBindService(tenants, &domainPanelStub{}, resolver, nil, config, repository.NewDomainBindJobRepository(database))
	restored, err := restarted.Status(ctx)
	if err != nil || restored.State != DomainStateVerified || restored.Domain != "restart.example.com" || restored.CurrentStep != 1 {
		t.Fatalf("Status(after restart) = %#v, %v", restored, err)
	}
}

func TestDomainUnbindDeletesSiteAndClearsDomain(t *testing.T) {
	tenants := domainBindTenants(t)
	customDomain := "academy.example.com"
	tenant := &domain.Tenant{Code: "unbind", Name: "Unbind", Status: 1, CustomDomain: &customDomain}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	panel := &domainPanelStub{}
	audit := &domainAuditStub{}
	svc := NewDomainBindService(tenants, panel, domainResolverStub{}, audit, DomainBindConfig{})
	status, err := svc.Unbind(ctx)
	if err != nil || status.State != DomainStateNone {
		t.Fatalf("Unbind() = %#v, %v", status, err)
	}
	if want := []string{"delete"}; !reflect.DeepEqual(panel.callNames(), want) {
		t.Fatalf("panel calls = %#v, want %#v", panel.callNames(), want)
	}
	stored, err := tenants.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.CustomDomain != nil {
		t.Fatalf("stored tenant = %#v, %v", stored, err)
	}
	if len(audit.events) != 1 || audit.events[0].Action != "domain.unbind" {
		t.Fatalf("audit events = %#v", audit.events)
	}
}

func TestDomainUnbindReservedDomainDoesNotDeletePanelSite(t *testing.T) {
	tenants := domainBindTenants(t)
	reservedDomain := "PLAY.IMAI.WORK."
	tenant := &domain.Tenant{
		Code:         "reserved",
		Name:         "Reserved",
		Status:       1,
		CustomDomain: &reservedDomain,
	}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(
		context.Background(),
		"admin",
		tenant.ID,
		"admin@example.com",
		"tenant_admin",
	)
	panel := &domainPanelStub{}
	service := NewDomainBindService(
		tenants,
		panel,
		domainResolverStub{},
		nil,
		DomainBindConfig{ReservedDomain: "play.imai.work"},
	)

	status, err := service.Unbind(ctx)
	if err != nil || status.State != DomainStateNone {
		t.Fatalf("Unbind() = %#v, %v", status, err)
	}
	if calls := panel.callNames(); len(calls) != 0 {
		t.Fatalf("panel calls = %#v, want none", calls)
	}
	stored, err := tenants.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.CustomDomain != nil {
		t.Fatalf("stored tenant = %#v, %v", stored, err)
	}
}

func waitDomainState(
	t *testing.T,
	svc *DomainBindService,
	ctx context.Context,
	want string,
) DomainBindStatus {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		status, err := svc.Status(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == want {
			return status
		}
		time.Sleep(time.Millisecond)
	}
	status, _ := svc.Status(ctx)
	t.Fatalf("status = %#v, want %q", status, want)
	return DomainBindStatus{}
}

func domainBindTenants(t *testing.T) repository.TenantRepository {
	t.Helper()
	database, tenants, _ := serviceRepositories(t)
	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatalf("get sqlite database: %v", err)
	}
	sqlDatabase.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDatabase.Close()
	})
	return tenants
}
