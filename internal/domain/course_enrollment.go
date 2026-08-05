package domain

const (
	AssignmentRequired = "required"
	AssignmentOptional = "optional"
)

type CourseEnrollment struct {
	BaseModel
	CourseID       string `gorm:"index;not null" json:"course_id"`
	UserID         string `gorm:"index;not null" json:"user_id"`
	Status         int    `gorm:"default:1" json:"status"`
	AssignmentType string `gorm:"not null;default:required" json:"assignment_type"`
}
