package domain

type Course struct {
	BaseModel
	Title         string  `gorm:"not null" json:"title"`
	Description   string  `json:"description"`
	CoverImage    string  `json:"cover_image"`
	Status        int     `gorm:"default:0" json:"status"`
	CreatedBy     string  `gorm:"index;not null" json:"created_by"`
	IsOfficial    bool    `gorm:"index;default:false" json:"is_official"`
	OwnerTenantID *string `gorm:"index" json:"owner_tenant_id,omitempty"`
	CategoryID    *string `gorm:"index" json:"category_id,omitempty"`
	Enabled       bool    `gorm:"->;-:migration" json:"enabled"`
}

type TenantOfficialCourse struct {
	TenantID string `gorm:"primaryKey;not null" json:"tenant_id"`
	CourseID string `gorm:"primaryKey;not null" json:"course_id"`
	Enabled  bool   `gorm:"not null;default:false" json:"enabled"`
}
