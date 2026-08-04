package domain

type CourseMaterial struct {
	BaseModel
	CourseID    string   `gorm:"index;not null" json:"course_id"`
	ResourceID  string   `gorm:"index;not null" json:"resource_id"`
	DisplayName string   `gorm:"not null" json:"display_name"`
	SortOrder   int      `gorm:"default:0" json:"sort_order"`
	CreatedBy   string   `gorm:"not null" json:"created_by"`
	Resource    Resource `gorm:"foreignKey:ResourceID" json:"resource"`
}
