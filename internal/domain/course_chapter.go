package domain

type CourseChapter struct {
	BaseModel
	CourseID  string `gorm:"index;not null" json:"course_id"`
	Title     string `gorm:"not null" json:"title"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
}
