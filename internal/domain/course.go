package domain

type Course struct {
	BaseModel
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	Status      int    `gorm:"default:0" json:"status"`
	CreatedBy   string `gorm:"index;not null" json:"created_by"`
}
