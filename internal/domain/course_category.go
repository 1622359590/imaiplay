package domain

type CourseCategory struct {
	BaseModel
	Name           string `gorm:"not null" json:"name"`
	NormalizedName string `gorm:"not null" json:"-"`
	SortOrder      int    `gorm:"default:0" json:"sort_order"`
	Status         int    `gorm:"default:1" json:"status"`
}
