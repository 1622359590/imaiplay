package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

func TestDashboardHandlerReturnsExactRoleShapes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name, role, tenantID string
		stats                interface{}
		wantKeys             []string
	}{
		{
			name: "tenant", role: "tenant_admin", tenantID: "tenant-1",
			stats: service.TenantDashboard{Scope: "tenant"},
			wantKeys: []string{
				"course_count", "has_demo_data", "learner_count", "manager_count",
				"published_course_count", "resource_category_count", "resource_count",
				"resource_type_counts", "scope", "today_learning_ranking",
				"today_learning_user_count", "today_learning_user_delta",
				"today_new_learner_count", "yesterday_learning_user_count",
			},
		},
		{
			name: "instructor", role: "instructor", tenantID: "tenant-1",
			stats: service.InstructorDashboard{Scope: "instructor"},
			wantKeys: []string{
				"course_count", "published_course_count", "recent_courses", "scope", "today_learning_user_count",
			},
		},
		{
			name: "platform", role: "superadmin", tenantID: "",
			stats: service.PlatformDashboard{Scope: "platform"},
			wantKeys: []string{
				"active_tenant_count", "course_count", "learner_count", "recent_tenants", "scope", "tenant_count",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewDashboardHandler(dashboardServiceStub{stats: test.stats})
			router := gin.New()
			router.Use(asUser(test.role, test.tenantID, "manager"))
			router.GET("/dashboard", handler.Get)
			response := requestJSON(t, router, http.MethodGet, "/dashboard", "")
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var body struct {
				Code int                    `json:"code"`
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Code != 0 {
				t.Fatalf("body=%s error=%v", response.Body.String(), err)
			}
			gotKeys := make([]string, 0, len(body.Data))
			for key := range body.Data {
				gotKeys = append(gotKeys, key)
			}
			sort.Strings(gotKeys)
			sort.Strings(test.wantKeys)
			if !equalStrings(gotKeys, test.wantKeys) {
				t.Fatalf("data keys=%#v want=%#v body=%s", gotKeys, test.wantKeys, response.Body.String())
			}
		})
	}
}

func TestDashboardHandlerTenantResourceTypesAlwaysContainFourKeys(t *testing.T) {
	handler := NewDashboardHandler(dashboardServiceStub{stats: service.TenantDashboard{Scope: "tenant"}})
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "manager"))
	router.GET("/dashboard", handler.Get)
	response := requestJSON(t, router, http.MethodGet, "/dashboard", "")
	var body struct {
		Data struct {
			ResourceTypes map[string]int64 `json:"resource_type_counts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(body.Data.ResourceTypes))
	for key := range body.Data.ResourceTypes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if !equalStrings(keys, []string{"attachment", "document", "image", "video"}) {
		t.Fatalf("resource type keys = %#v body=%s", keys, response.Body.String())
	}
}

func TestDashboardHandlerRejectsUnauthorizedRoles(t *testing.T) {
	handler := NewDashboardHandler(dashboardServiceStub{})
	router := gin.New()
	router.Use(asUser("learner", "tenant-1", "learner"))
	router.GET("/dashboard", handler.Get)
	response := requestJSON(t, router, http.MethodGet, "/dashboard", "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type dashboardServiceStub struct {
	stats interface{}
	err   error
}

func (stub dashboardServiceStub) Stats(context.Context) (interface{}, error) {
	return stub.stats, stub.err
}
