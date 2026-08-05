package api

import (
	"net/http"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCourseCategoryHandlerTenantCRUDAndReferencedConflict(t *testing.T) {
	categoryService, database := courseCategoryHandlerService(t)
	router := courseCategoryTestRouter(categoryService, "tenant_admin", "tenant-1")
	created := requestJSON(t, router, http.MethodPost, "/course-categories", `{"name":"  Sales  ","sort_order":2,"status":1}`)
	if created.Code != http.StatusOK {
		t.Fatalf("Create status=%d body=%s", created.Code, created.Body.String())
	}
	id := responseID(t, created.Body.Bytes())
	if response := requestJSON(t, router, http.MethodGet, "/course-categories", ""); response.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, router, http.MethodPut, "/course-categories/"+id, `{"name":"Revenue","sort_order":3,"status":1}`); response.Code != http.StatusOK {
		t.Fatalf("Update status=%d body=%s", response.Code, response.Body.String())
	}
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Title:     "Referenced", CreatedBy: "admin-1", CategoryID: &id,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatal(err)
	}
	response := requestJSON(t, router, http.MethodDelete, "/course-categories/"+id, "")
	if response.Code != http.StatusConflict {
		t.Fatalf("Delete referenced status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCourseCategoryHandlerRoleAndPlatformBoundaries(t *testing.T) {
	categoryService, _ := courseCategoryHandlerService(t)
	instructor := courseCategoryTestRouter(categoryService, "instructor", "tenant-1")
	if response := requestJSON(t, instructor, http.MethodGet, "/course-categories", ""); response.Code != http.StatusOK {
		t.Fatalf("instructor List status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, instructor, http.MethodPost, "/course-categories", `{"name":"Forbidden"}`); response.Code != http.StatusForbidden {
		t.Fatalf("instructor Create status=%d body=%s", response.Code, response.Body.String())
	}
	admin := courseCategoryTestRouter(categoryService, "tenant_admin", "tenant-1")
	if response := requestJSON(t, admin, http.MethodGet, "/admin/course-categories", ""); response.Code != http.StatusForbidden {
		t.Fatalf("tenant admin platform List status=%d body=%s", response.Code, response.Body.String())
	}
	root := courseCategoryTestRouter(categoryService, "superadmin", "")
	created := requestJSON(t, root, http.MethodPost, "/admin/course-categories", `{"name":"Official"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("platform Create status=%d body=%s", created.Code, created.Body.String())
	}
	if response := requestJSON(t, root, http.MethodGet, "/admin/course-categories", ""); response.Code != http.StatusOK {
		t.Fatalf("platform List status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, root, http.MethodGet, "/course-categories", ""); response.Code != http.StatusForbidden {
		t.Fatalf("superadmin tenant List status=%d body=%s", response.Code, response.Body.String())
	}
}

func courseCategoryHandlerService(t *testing.T) (*service.CourseCategoryService, *gorm.DB) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	return service.NewCourseCategoryService(repository.NewCourseCategoryRepository(database)), database
}

func courseCategoryTestRouter(categoryService *service.CourseCategoryService, role, tenantID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewCourseCategoryHandler(categoryService)
	router := gin.New()
	router.Use(asUser(role, tenantID, "user-1"))
	router.GET("/course-categories", handler.List)
	router.POST("/course-categories", handler.Create)
	router.PUT("/course-categories/:id", handler.Update)
	router.DELETE("/course-categories/:id", handler.Delete)
	router.GET("/admin/course-categories", handler.ListPlatform)
	router.POST("/admin/course-categories", handler.CreatePlatform)
	router.PUT("/admin/course-categories/:id", handler.UpdatePlatform)
	router.DELETE("/admin/course-categories/:id", handler.DeletePlatform)
	return router
}
