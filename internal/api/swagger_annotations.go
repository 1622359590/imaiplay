package api

// Health documents the service health endpoint.
// @Summary Service health
// @Tags system
// @Produce json
// @Success 200 {object} APIResponse
// @Router /health [get]
func Health() {}

// DatabaseHealth documents the database health endpoint.
// @Summary Database health
// @Tags system
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 503 {object} APIResponse
// @Router /health/db [get]
func DatabaseHealth() {}

// Register documents public account registration.
// @Summary Register an account
// @Tags auth
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Registration request"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Router /api/v1/auth/register [post]
func Register() {}

// Login documents account login.
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Login request"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /api/v1/auth/login [post]
func Login() {}

// Refresh documents token refresh.
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Refresh request"
// @Success 200 {object} APIResponse
// @Router /api/v1/auth/refresh [post]
func Refresh() {}

// ListCourses documents course listing.
// @Summary List courses
// @Tags courses
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Page size"
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/courses [get]
func ListCourses() {}

// GetCourse documents course details.
// @Summary Get course
// @Tags courses
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/courses/{id} [get]
func GetCourse() {}

// CreateCourse documents course creation.
// @Summary Create course
// @Tags courses
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Course"
// @Success 201 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/courses [post]
func CreateCourse() {}

// UpdateCourse documents course updates.
// @Summary Update course
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Param body body map[string]interface{} true "Course"
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/courses/{id} [put]
func UpdateCourse() {}

// DeleteCourse documents course deletion.
// @Summary Delete course
// @Tags courses
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/courses/{id} [delete]
func DeleteCourse() {}

// UploadResource documents resource uploads.
// @Summary Upload a resource
// @Tags resources
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Resource file"
// @Success 201 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/resources [post]
func UploadResource() {}

// ReportProgress documents lesson progress reporting.
// @Summary Report lesson progress
// @Tags learning
// @Accept json
// @Produce json
// @Param id path string true "Lesson ID"
// @Param body body map[string]interface{} true "Progress"
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/lessons/{id}/progress [post]
func ReportProgress() {}

// Dashboard documents dashboard statistics.
// @Summary Dashboard statistics
// @Tags dashboard
// @Produce json
// @Success 200 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/dashboard [get]
func Dashboard() {}
