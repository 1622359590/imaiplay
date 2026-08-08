package domain

type Resource struct {
	BaseModel
	CategoryID      *string `json:"category_id,omitempty"`
	Name            string  `gorm:"not null" json:"name"`
	ResourceType    string  `gorm:"not null" json:"resource_type"`
	URL             string  `gorm:"not null" json:"url"`
	SizeBytes       int64   `gorm:"default:0" json:"size_bytes"`
	DurationSeconds int     `gorm:"default:0" json:"duration_seconds"`
	CreatedBy       string  `gorm:"not null" json:"created_by"`
}
