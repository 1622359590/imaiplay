package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEnrollmentHandlerTenantAdminFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	adminContext := asUser("tenant_admin", tenant.ID, "admin-1")
	learner, err := services.users.Create(
		withRole("tenant_admin", tenant.ID, "admin-1"),
		"learner@example.com", "password123", "Learner", "learner",
	)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	course, err := services.courses.Create(
		withRole("tenant_admin", tenant.ID, "admin-1"), "Course", "", "",
	)
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	handler := NewEnrollmentHandler(services.enrollments)
	router := gin.New()
	router.Use(adminContext)
	router.POST("/courses/:id/enrollments", handler.Enroll)
	router.GET("/courses/:id/enrollments", handler.ListByCourse)
	router.DELETE("/enrollments/:id", handler.Remove)

	created := requestJSON(t, router, http.MethodPost,
		"/courses/"+course.ID+"/enrollments",
		`{"user_id":"`+learner.ID+`"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("Enroll status=%d body=%s", created.Code, created.Body.String())
	}
	enrollmentID := responseID(t, created.Body.Bytes())
	list := requestJSON(t, router, http.MethodGet,
		"/courses/"+course.ID+"/enrollments", "")
	if list.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", list.Code, list.Body.String())
	}
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &body); err != nil ||
		len(body.Data) != 1 || body.Data[0].ID != enrollmentID {
		t.Fatalf("List body=%s error=%v", list.Body.String(), err)
	}
	if response := requestJSON(t, router, http.MethodDelete,
		"/enrollments/"+enrollmentID, ""); response.Code != http.StatusOK {
		t.Fatalf("Remove status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestEnrollmentHandlerRejectsNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewEnrollmentHandler(services.enrollments)
	router := gin.New()
	router.Use(asUser("instructor", "tenant-1", "instructor-1"))
	router.POST("/courses/:id/enrollments", handler.Enroll)
	response := requestJSON(
		t, router, http.MethodPost, "/courses/course-1/enrollments",
		`{"user_id":"learner-1"}`,
	)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
