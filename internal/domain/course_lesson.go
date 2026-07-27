package domain

type CourseLesson struct {
	BaseModel
	ChapterID       string  `gorm:"index;not null" json:"chapter_id"`
	Title           string  `gorm:"not null" json:"title"`
	ContentType     string  `gorm:"not null" json:"content_type"`
	ResourceID      *string `gorm:"index" json:"resource_id,omitempty"`
	ResourceType    string  `gorm:"-" json:"resource_type,omitempty"`
	ContentURL      string  `json:"content_url"`
	DurationSeconds int     `gorm:"default:0" json:"duration_seconds"`
	SortOrder       int     `gorm:"default:0" json:"sort_order"`
}
