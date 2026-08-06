package server

import "github.com/gin-gonic/gin"

func registerCourseRoutes(backend *gin.RouterGroup, h routeHandlers) {
	backend.POST("/courses", h.course.Create)
	backend.POST("/official-courses", h.course.CreateOfficial)
	backend.GET("/official-courses", h.course.OfficialList)
	backend.PUT("/official-courses/:id/enabled", h.course.EnableOfficial)
	backend.GET("/courses", h.course.List)
	backend.GET("/courses/:id", h.course.Get)
	backend.PUT("/courses/:id", h.course.Update)
	backend.DELETE("/courses/:id", h.course.Delete)
	backend.GET("/courses/:id/detail", h.course.Detail)
	backend.GET("/courses/:id/materials", h.material.List)
	backend.POST("/courses/:id/materials", h.material.Add)
	backend.PUT("/courses/:id/materials/:materialID", h.material.Update)
	backend.DELETE("/courses/:id/materials/:materialID", h.material.Remove)
	backend.POST("/courses/:id/chapters", h.chapter.Create)
	backend.GET("/courses/:id/chapters", h.chapter.List)
	backend.PUT("/chapters/:id", h.chapter.Update)
	backend.DELETE("/chapters/:id", h.chapter.Delete)
	backend.POST("/chapters/:id/lessons", h.lesson.Create)
	backend.GET("/chapters/:id/lessons", h.lesson.List)
	backend.PUT("/lessons/:id", h.lesson.Update)
	backend.DELETE("/lessons/:id", h.lesson.Delete)
	backend.POST("/courses/:id/enrollments", h.enrollment.Enroll)
	backend.GET("/courses/:id/enrollments", h.enrollment.ListByCourse)
	backend.PUT("/enrollments/:id", h.enrollment.UpdateAssignment)
	backend.DELETE("/enrollments/:id", h.enrollment.Remove)
}
