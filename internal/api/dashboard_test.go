package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

func TestDashboardHandlerReturnsStatsForManagerRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, role := range []string{"tenant_admin", "instructor"} {
		t.Run(role, func(t *testing.T) {
			handler := NewDashboardHandler(dashboardServiceStub{
				stats: service.DashboardStats{
					UserCount: 10, CourseCount: 5,
					PublishedCourseCount: 4, TodayNewUserCount: 2,
					TodayLearningUserCount: 3, TotalLearningSeconds: 3600,
					CourseCompletionRate: 0.75,
				},
			})
			router := gin.New()
			router.Use(asUser(role, "tenant-1", "manager"))
			router.GET("/dashboard", handler.Get)
			response := requestJSON(
				t, router, http.MethodGet, "/dashboard", "",
			)
			if response.Code != http.StatusOK {
				t.Fatalf(
					"status=%d body=%s",
					response.Code, response.Body.String(),
				)
			}
			var body struct {
				Code int                    `json:"code"`
				Data service.DashboardStats `json:"data"`
			}
			want := service.DashboardStats{
				UserCount: 10, CourseCount: 5,
				PublishedCourseCount: 4, TodayNewUserCount: 2,
				TodayLearningUserCount: 3, TotalLearningSeconds: 3600,
				CourseCompletionRate: 0.75,
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil ||
				body.Code != 0 || body.Data != want {
				t.Fatalf("body=%s error=%v", response.Body.String(), err)
			}
		})
	}
}

func TestDashboardHandlerRejectsUnauthorizedRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, role := range []string{"learner", "superadmin"} {
		t.Run(role, func(t *testing.T) {
			handler := NewDashboardHandler(dashboardServiceStub{})
			router := gin.New()
			router.Use(asUser(role, "tenant-1", "user"))
			router.GET("/dashboard", handler.Get)
			response := requestJSON(
				t, router, http.MethodGet, "/dashboard", "",
			)
			if response.Code != http.StatusForbidden {
				t.Fatalf(
					"status=%d body=%s",
					response.Code, response.Body.String(),
				)
			}
		})
	}
}

type dashboardServiceStub struct {
	stats service.DashboardStats
	err   error
}

func (stub dashboardServiceStub) Stats(
	context.Context,
) (service.DashboardStats, error) {
	return stub.stats, stub.err
}
