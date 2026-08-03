package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type domainBindServiceStub struct {
	verified int
	bound    int
	statused int
	unbound  int
}

func (stub *domainBindServiceStub) Verify(
	context.Context,
	string,
) (service.DomainBindStatus, error) {
	stub.verified++
	return service.DomainBindStatus{
		State: service.DomainStateVerified, Domain: "academy.example.com",
	}, nil
}

func (stub *domainBindServiceStub) Bind(
	context.Context,
	string,
) (service.DomainBindStatus, error) {
	stub.bound++
	return service.DomainBindStatus{
		State: service.DomainStateCreatingSite, Domain: "academy.example.com",
	}, nil
}

func (stub *domainBindServiceStub) Status(
	context.Context,
) (service.DomainBindStatus, error) {
	stub.statused++
	return service.DomainBindStatus{
		State: service.DomainStateReady, Domain: "academy.example.com",
		TenantCode: "acme", DefaultPortalURL: "https://play.imai.work/t/acme",
	}, nil
}

func (stub *domainBindServiceStub) Unbind(
	context.Context,
) (service.DomainBindStatus, error) {
	stub.unbound++
	return service.DomainBindStatus{State: service.DomainStateNone}, nil
}

func TestDomainBindHandlerRoutesAndRoleCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &domainBindServiceStub{}
	handler := NewDomainBindHandler(stub)
	router := gin.New()
	router.Use(asRole("tenant_admin", "tenant-1"))
	router.POST("/domain-bind/verify", handler.Verify)
	router.POST("/domain-bind", handler.Bind)
	router.GET("/domain-bind/status", handler.Status)
	router.DELETE("/domain-bind", handler.Unbind)

	for _, request := range []struct {
		method, path, body string
	}{
		{http.MethodPost, "/domain-bind/verify", `{"domain":"academy.example.com"}`},
		{http.MethodPost, "/domain-bind", `{"domain":"academy.example.com"}`},
		{http.MethodGet, "/domain-bind/status", ""},
		{http.MethodDelete, "/domain-bind", ""},
	} {
		response := requestJSON(t, router, request.method, request.path, request.body)
		if response.Code != http.StatusOK {
			t.Fatalf("%s %s status=%d body=%s", request.method, request.path, response.Code, response.Body.String())
		}
	}
	if stub.verified != 1 || stub.bound != 1 || stub.statused != 1 || stub.unbound != 1 {
		t.Fatalf("service calls = %#v", stub)
	}

	invalid := requestJSON(t, router, http.MethodPost, "/domain-bind/verify", `{}`)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid status=%d body=%s", invalid.Code, invalid.Body.String())
	}

	forbidden := gin.New()
	forbidden.Use(asRole("learner", "tenant-1"))
	forbidden.POST("/domain-bind/verify", handler.Verify)
	response := requestJSON(
		t, forbidden, http.MethodPost, "/domain-bind/verify",
		`{"domain":"academy.example.com"}`,
	)
	if response.Code != http.StatusForbidden {
		t.Fatalf("forbidden status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestDomainBindStatusReturnsDefaultPortalMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &domainBindServiceStub{}
	handler := NewDomainBindHandler(stub)
	router := gin.New()
	router.Use(asRole("tenant_admin", "tenant-1"))
	router.GET("/domain-bind/status", handler.Status)

	response := requestJSON(t, router, http.MethodGet, "/domain-bind/status", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"tenant_code":"acme"`) ||
		!strings.Contains(response.Body.String(), `"default_portal_url":"https://play.imai.work/t/acme"`) {
		t.Fatalf("response = %s", response.Body.String())
	}
}
