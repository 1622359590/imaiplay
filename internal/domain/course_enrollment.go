package domain

type CourseEnrollment struct {
	BaseModel
	CourseID string `gorm:"index;not null" json:"course_id"`
	UserID   string `gorm:"index;not null" json:"user_id"`
	Status   int    `gorm:"default:1" json:"status"`
}
